package builders

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/sklinkert/go-ddd/internal/application/services"
	rest "github.com/sklinkert/go-ddd/internal/handler/api/rest"
	"github.com/sklinkert/go-ddd/internal/handler/api/rest/httpx"
	restmiddleware "github.com/sklinkert/go-ddd/internal/handler/api/rest/middleware"
	db "github.com/sklinkert/go-ddd/internal/infrastructure/db/postgres"
	"github.com/sklinkert/go-ddd/internal/infrastructure/logging"
	"go.uber.org/zap"
)

const (
	defaultServeDSN         = "host=localhost user=marketplace password=marketplace dbname=marketplace port=5432 sslmode=disable"
	defaultServePort        = ":8080"
	defaultServeEnvironment = "development"
	defaultServeLogLevel    = "debug"
)

type ServerBuilder struct {
	dsn         string
	port        string
	corsOrigins []string
	environment string
	logLevel    string
}

type ServerApp struct {
	app    *fiber.App
	conn   *pgx.Conn
	port   string
	logger *zap.Logger
}

func NewServerBuilder() *ServerBuilder {
	return &ServerBuilder{
		dsn:         defaultServeDSN,
		port:        defaultServePort,
		corsOrigins: []string{"*"},
		environment: defaultServeEnvironment,
		logLevel:    defaultServeLogLevel,
	}
}

func (b *ServerBuilder) WithDSN(dsn string) *ServerBuilder {
	if strings.TrimSpace(dsn) != "" {
		b.dsn = strings.TrimSpace(dsn)
	}
	return b
}

func (b *ServerBuilder) WithPort(port string) *ServerBuilder {
	if strings.TrimSpace(port) == "" {
		return b
	}

	normalized := strings.TrimSpace(port)
	if !strings.HasPrefix(normalized, ":") {
		normalized = ":" + normalized
	}
	b.port = normalized

	return b
}

func (b *ServerBuilder) WithCORSOrigins(origins []string) *ServerBuilder {
	if len(origins) == 0 {
		return b
	}

	filtered := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}

	if len(filtered) > 0 {
		b.corsOrigins = filtered
	}

	return b
}

func (b *ServerBuilder) WithEnvironment(environment string) *ServerBuilder {
	if strings.TrimSpace(environment) != "" {
		b.environment = strings.TrimSpace(environment)
	}
	return b
}

func (b *ServerBuilder) WithLogLevel(logLevel string) *ServerBuilder {
	if strings.TrimSpace(logLevel) != "" {
		b.logLevel = strings.TrimSpace(logLevel)
	}
	return b
}

func (b *ServerBuilder) Build(ctx context.Context) (*ServerApp, error) {
	logger, err := logging.NewLogger(logging.Config{
		Environment: b.environment,
		Level:       b.logLevel,
	})
	if err != nil {
		return nil, err
	}

	var dbTraceLogger *zap.Logger
	if logging.IsTraceLevel(b.logLevel) {
		dbTraceLogger = logger
	}

	conn, err := db.NewConnectionWithLogger(ctx, b.dsn, dbTraceLogger)
	if err != nil {
		_ = logger.Sync()
		return nil, err
	}

	queries := db.NewQueries(conn)
	productRepo := db.NewSqlcProductRepository(queries)
	sellerRepo := db.NewSqlcSellerRepository(queries)
	idempotencyRepo := db.NewSqlcIdempotencyRepository(queries)

	productService := services.NewProductService(productRepo, sellerRepo, idempotencyRepo)
	sellerService := services.NewSellerService(sellerRepo, idempotencyRepo)

	app := fiber.New(fiber.Config{
		AppName:               "CQRS API Server",
		BodyLimit:             2 * 1024 * 1024, // 2MB
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.Error("unhandled request error",
				zap.String("method", c.Method()),
				zap.String("path", c.OriginalURL()),
				zap.Error(err),
			)
			return httpx.WriteError(c, err)
		},
	})
	restmiddleware.Register(app, restmiddleware.Config{
		Logger:      logger,
		Environment: b.environment,
		CORSOrigins: b.corsOrigins,
	})

	rest.NewHealthController(app, b.environment, conn, logger)
	rest.NewProductController(app, productService)
	rest.NewSellerController(app, sellerService)

	return &ServerApp{
		app:    app,
		conn:   conn,
		port:   b.port,
		logger: logger,
	}, nil
}

func (a *ServerApp) Start() error {
	a.logger.Info("starting HTTP server", zap.String("port", a.port))
	return a.app.Listen(a.port)
}

func (a *ServerApp) Shutdown(ctx context.Context) error {
	if a.app == nil {
		return nil
	}
	return a.app.ShutdownWithContext(ctx)
}

func (a *ServerApp) Close(ctx context.Context) error {
	if a.conn == nil {
		if a.logger != nil {
			_ = a.logger.Sync()
		}
		return nil
	}
	a.conn.Close(ctx)
	if a.logger != nil {
		_ = a.logger.Sync()
	}
	return nil
}
