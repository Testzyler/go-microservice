package proxy

import (
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	authv1connect "github.com/Testzyler/go-microservice/gen/proto/auth/v1/authv1connect"
	"github.com/Testzyler/go-microservice/global/pkg/token"
	"github.com/Testzyler/go-microservice/services/internal-gateway/internal/config"
	"github.com/Testzyler/go-microservice/services/internal-gateway/internal/middleware"
)

// Router handles routing and proxying requests to backend services
type Router struct {
	app          *fiber.App
	config       *config.Config
	logger       *zap.Logger
	tokenService *token.JWTService
	authClient   authv1connect.AuthServiceClient
}

// NewRouter creates a new router with minimal auth-aware routes
func NewRouter(cfg *config.Config, logger *zap.Logger) (*Router, error) {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	})

	tokenService := token.NewJWTService(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTokenTTL)
	httpClient := &http.Client{Timeout: 10 * time.Second}
	authClient := authv1connect.NewAuthServiceClient(httpClient, cfg.AuthServiceURL, connect.WithGRPC())

	router := &Router{
		app:          app,
		config:       cfg,
		logger:       logger,
		tokenService: tokenService,
		authClient:   authClient,
	}

	router.setupMiddleware()
	router.setupRoutes()

	return router, nil
}

func (r *Router) setupMiddleware() {
	// OpenTelemetry tracing middleware - creates root span for each request
	if r.config.EnableMonitoring {
		r.app.Use(otelfiber.Middleware(
			otelfiber.WithServerName("internal-gateway"),
			otelfiber.WithSpanNameFormatter(func(c *fiber.Ctx) string {
				return c.Method() + " " + c.Path()
			}),
		))
	}

	// CORS
	r.app.Use(middleware.NewCORSMiddleware(middleware.CORSConfig{
		AllowOrigins: r.config.CORSAllowOrigins,
		AllowMethods: r.config.CORSAllowMethods,
		AllowHeaders: r.config.CORSAllowHeaders,
	}))

	// Rate limiting
	if r.config.RateLimitEnabled {
		rateLimiter := middleware.NewRateLimiter(r.config.RateLimitRPS, r.config.RateLimitBurst)
		r.app.Use(rateLimiter.Middleware())
	}

	// Request logging with trace context
	r.app.Use(func(c *fiber.Ctx) error {
		span := trace.SpanFromContext(c.UserContext())
		sc := span.SpanContext()
		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("ip", c.IP()),
		}
		if r.config.EnableMonitoring && sc.IsValid() {
			fields = append(fields,
				zap.String("trace_id", sc.TraceID().String()),
				zap.String("span_id", sc.SpanID().String()),
			)
		}
		r.logger.Info("incoming request", fields...)
		return c.Next()
	})
}

func (r *Router) setupRoutes() {
	r.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "internal-gateway",
		})
	})

	authGroup := r.app.Group("/api/v1/auth")
	h := newAuthHandlers(r.authClient, r.tokenService, r.logger, r.config)
	authGroup.Post("/register", h.register)
	authGroup.Post("/login", h.login)
	authGroup.Post("/token", h.issueToken)
	authGroup.Get("/me", h.authValidate, h.getCurrentUser)
}

// Start starts the HTTP server
func (r *Router) Start() error {
	r.logger.Info("starting internal-gateway", zap.String("port", r.config.Port))
	return r.app.Listen(":" + r.config.Port)
}

// Shutdown gracefully shuts down the router
func (r *Router) Shutdown() error {
	return r.app.Shutdown()
}

func writeConnectError(c *fiber.Ctx, err error) error {
	if connectErr, ok := err.(*connect.Error); ok {
		status := fiber.StatusInternalServerError
		switch connectErr.Code() {
		case connect.CodeInvalidArgument:
			status = fiber.StatusBadRequest
		case connect.CodeUnauthenticated:
			status = fiber.StatusUnauthorized
		case connect.CodeAlreadyExists:
			status = fiber.StatusConflict
		default:
			status = fiber.StatusInternalServerError
		}
		return c.Status(status).JSON(fiber.Map{"error": connectErr.Message()})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
}
