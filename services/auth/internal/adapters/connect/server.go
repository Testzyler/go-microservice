package connect

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	authv1 "github.com/Testzyler/go-microservice/gen/proto/auth/v1"
	authv1connect "github.com/Testzyler/go-microservice/gen/proto/auth/v1/authv1connect"
	"github.com/Testzyler/go-microservice/services/auth/internal/application"
	"github.com/Testzyler/go-microservice/services/auth/internal/domain/user"
)

type Server struct {
	authv1connect.UnimplementedAuthServiceHandler
	auth    *application.AuthService
	log     *zap.Logger
	http    *http.Server
	metrics http.Handler
}

func NewServer(auth *application.AuthService, log *zap.Logger, metrics http.Handler) *Server {
	return &Server{auth: auth, log: log, metrics: metrics}
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
	interceptors := connect.WithInterceptors(otelInterceptor, logInterceptor)

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
		Addr:    addr,
		Handler: otelhttp.NewHandler(h2c.NewHandler(mux, &http2.Server{}), "auth-connect"),
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
	result, err := s.auth.Register(ctx, req.Msg.Email, req.Msg.Password)
	if err != nil {
		return nil, connect.NewError(codeFor(err), err)
	}
	return connect.NewResponse(&authv1.RegisterResponse{
		UserId:      result.UserID.String(),
		AccessToken: result.AccessToken,
		Roles:       result.Roles,
		Permissions: result.Permissions,
		ExpiresIn:   result.ExpiresIn,
	}), nil
}

func (s *Server) Login(ctx context.Context, req *connect.Request[authv1.LoginRequest]) (*connect.Response[authv1.LoginResponse], error) {
	result, err := s.auth.Login(ctx, req.Msg.Email, req.Msg.Password)
	if err != nil {
		return nil, connect.NewError(codeFor(err), err)
	}
	return connect.NewResponse(&authv1.LoginResponse{
		UserId:      result.UserID.String(),
		AccessToken: result.AccessToken,
		Roles:       result.Roles,
		Permissions: result.Permissions,
		ExpiresIn:   result.ExpiresIn,
	}), nil
}

func (s *Server) Validate(ctx context.Context, req *connect.Request[authv1.ValidateRequest]) (*connect.Response[authv1.ValidateResponse], error) {
	claims, err := s.auth.Validate(ctx, req.Msg.Token)
	if err != nil {
		return connect.NewResponse(&authv1.ValidateResponse{Valid: false, Reason: "invalid token"}), nil
	}
	return connect.NewResponse(&authv1.ValidateResponse{
		Valid:       true,
		UserId:      claims.Subject.String(),
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}), nil
}

func codeFor(err error) connect.Code {
	switch {
	case errorsIs(err, user.ErrUserExists):
		return connect.CodeAlreadyExists
	case errorsIs(err, user.ErrBadPassword), errorsIs(err, user.ErrNotFound):
		return connect.CodeUnauthenticated
	case errorsIs(err, user.ErrInvalidEmail):
		return connect.CodeInvalidArgument
	default:
		return connect.CodeInternal
	}
}

func errorsIs(err, target error) bool {
	return err != nil && target != nil && errors.Is(err, target)
}
