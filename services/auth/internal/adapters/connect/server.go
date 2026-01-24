package connect

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/bufbuild/protovalidate-go"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	authv1 "github.com/Testzyler/go-microservice/gen/proto/auth/v1"
	authv1connect "github.com/Testzyler/go-microservice/gen/proto/auth/v1/authv1connect"
	"github.com/Testzyler/go-microservice/pkg/errx"
	"github.com/Testzyler/go-microservice/services/auth/internal/application"
)

type Server struct {
	authv1connect.UnimplementedAuthServiceHandler
	auth      *application.AuthService
	log       *zap.Logger
	http      *http.Server
	metrics   http.Handler
	validator *protovalidate.Validator
}

func NewServer(auth *application.AuthService, log *zap.Logger, metrics http.Handler) (*Server, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &Server{auth: auth, log: log, metrics: metrics, validator: validator}, nil
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	otelInterceptor, err := otelconnect.NewInterceptor()
	if err != nil {
		return err
	}
	logInterceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			s.log.Info("connect request", zap.String("procedure", req.Spec().Procedure))
			return next(ctx, req)
		}
	})
	errInterceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			resp, err := next(ctx, req)
			if err != nil {
				return nil, normalizeError(ctx, err)
			}
			return resp, nil
		}
	})
	interceptors := connect.WithInterceptors(errInterceptor, otelInterceptor, logInterceptor)

	path, handler := authv1connect.NewAuthServiceHandler(s, interceptors)
	mux.Handle(path, handler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})
	if s.metrics != nil {
		mux.Handle("/metrics", s.metrics)
	}

	s.http = &http.Server{
		Addr: addr,
		Handler: otelhttp.NewHandler(
			h2c.NewHandler(mux, &http2.Server{}),
			"auth-connect",
			otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
				route := httpRoute(r)
				if route == "" {
					return r.Method
				}
				return r.Method + " " + route
			}),
			otelhttp.WithMetricAttributesFn(func(r *http.Request) []attribute.KeyValue {
				route := httpRoute(r)
				if route == "" {
					return nil
				}
				return []attribute.KeyValue{attribute.String("http.route", route)}
			}),
		),
	}
	s.log.Info("starting auth connect server", zap.String("addr", addr))
	return s.http.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}

func (s *Server) Register(ctx context.Context, req *connect.Request[authv1.RegisterRequest]) (*connect.Response[authv1.RegisterResponse], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, errx.New("auth.validation_failed", "validation failed", errx.KindInvalidArgument).WithCause(err)
	}
	result, err := s.auth.Register(ctx, req.Msg.Email, req.Msg.Password)
	if err != nil {
		return nil, err
	}
	resp := connect.NewResponse(&authv1.RegisterResponse{
		UserId:      result.UserID.String(),
		AccessToken: result.AccessToken,
		Roles:       result.Roles,
		Permissions: result.Permissions,
		ExpiresIn:   result.ExpiresIn,
	})
	attachTraceID(ctx, resp.Header())
	return resp, nil
}

func (s *Server) Login(ctx context.Context, req *connect.Request[authv1.LoginRequest]) (*connect.Response[authv1.LoginResponse], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, errx.New("auth.validation_failed", "validation failed", errx.KindInvalidArgument).WithCause(err)
	}
	result, err := s.auth.Login(ctx, req.Msg.Email, req.Msg.Password)
	if err != nil {
		return nil, err
	}
	resp := connect.NewResponse(&authv1.LoginResponse{
		UserId:      result.UserID.String(),
		AccessToken: result.AccessToken,
		Roles:       result.Roles,
		Permissions: result.Permissions,
		ExpiresIn:   result.ExpiresIn,
	})
	attachTraceID(ctx, resp.Header())
	return resp, nil
}

func (s *Server) Validate(ctx context.Context, req *connect.Request[authv1.ValidateRequest]) (*connect.Response[authv1.ValidateResponse], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		return nil, errx.New("auth.validation_failed", "validation failed", errx.KindInvalidArgument).WithCause(err)
	}
	claims, err := s.auth.Validate(ctx, req.Msg.Token)
	if err != nil {
		resp := connect.NewResponse(&authv1.ValidateResponse{Valid: false, Reason: "invalid token"})
		attachTraceID(ctx, resp.Header())
		return resp, nil
	}
	resp := connect.NewResponse(&authv1.ValidateResponse{
		Valid:       true,
		UserId:      claims.Subject.String(),
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	})
	attachTraceID(ctx, resp.Header())
	return resp, nil
}

func normalizeError(ctx context.Context, err error) *connect.Error {
	if err == nil {
		return nil
	}
	var connErr *connect.Error
	if errors.As(err, &connErr) {
		if connErr.Meta().Get("x-error-code") == "" {
			connErr.Meta().Set("x-error-code", errx.Code(err))
		}
		attachTraceID(ctx, connErr.Meta())
		return connErr
	}
	code := connectCodeForKind(errx.KindOf(err))
	connErr = connect.NewError(code, err)
	connErr.Meta().Set("x-error-code", errx.Code(err))
	attachTraceID(ctx, connErr.Meta())
	return connErr
}

func connectCodeForKind(kind errx.Kind) connect.Code {
	switch kind {
	case errx.KindInvalidArgument:
		return connect.CodeInvalidArgument
	case errx.KindNotFound:
		return connect.CodeNotFound
	case errx.KindConflict:
		return connect.CodeAlreadyExists
	case errx.KindUnauthenticated:
		return connect.CodeUnauthenticated
	case errx.KindForbidden:
		return connect.CodePermissionDenied
	case errx.KindUnavailable:
		return connect.CodeUnavailable
	default:
		return connect.CodeInternal
	}
}

func attachTraceID(ctx context.Context, headers http.Header) {
	if headers == nil {
		return
	}
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}
	sc := span.SpanContext()
	if !sc.IsValid() {
		return
	}
	headers.Set("x-trace-id", sc.TraceID().String())
}

func httpRoute(r *http.Request) string {
	if r == nil {
		return ""
	}
	if r.URL == nil {
		return ""
	}
	if r.URL.Path != "" {
		return r.URL.Path
	}
	return r.Pattern
}
