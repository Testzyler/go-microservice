package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"
)

type Config struct {
	Logger      *zap.Logger
	Environment string
	CORSOrigins []string
}

func Register(app *fiber.App, cfg Config) {
	app.Use(requestid.New())
	app.Use(recover.New(recover.Config{
		EnableStackTrace: normalizeEnvironment(cfg.Environment) != "production",
	}))

	if cfg.Logger != nil {
		app.Use(RequestLogger(cfg.Logger, cfg.Environment))
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins: buildAllowOrigins(cfg.CORSOrigins),
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,Idempotency-Key,X-Request-ID",
		MaxAge:       86400,
	}))
	app.Use(helmet.New())
	app.Use(limiter.New(limiter.Config{
		Max:        rateLimitMax(cfg.Environment),
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}))
}

func rateLimitMax(environment string) int {
	if normalizeEnvironment(environment) == "production" {
		return 120
	}
	return 1000
}

func normalizeEnvironment(environment string) string {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "prod":
		return "production"
	case "dev":
		return "development"
	default:
		if strings.TrimSpace(environment) == "" {
			return "development"
		}
		return strings.ToLower(strings.TrimSpace(environment))
	}
}

func buildAllowOrigins(origins []string) string {
	if len(origins) == 0 {
		return "*"
	}

	filtered := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}

	if len(filtered) == 0 {
		return "*"
	}

	return strings.Join(filtered, ",")
}
