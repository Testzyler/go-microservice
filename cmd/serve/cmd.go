package serve

import (
	"context"
	"os"
	"strings"

	"github.com/sklinkert/go-ddd/internal/builders"
)

type Config struct {
	DSN         string
	Port        string
	CORSOrigins []string
	Environment string
	LogLevel    string
}

func DefaultConfig() Config {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		dsn = "host=localhost user=marketplace password=marketplace dbname=marketplace port=5432 sslmode=disable"
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = ":8080"
	}

	corsOrigins := []string{"*"}
	if raw := strings.TrimSpace(os.Getenv("CORS_ORIGINS")); raw != "" {
		corsOrigins = splitCSV(raw)
		if len(corsOrigins) == 0 {
			corsOrigins = []string{"*"}
		}
	}

	environment := strings.TrimSpace(os.Getenv("APP_ENV"))
	if environment == "" {
		environment = "development"
	}

	logLevel := strings.TrimSpace(os.Getenv("LOG_LEVEL"))
	if logLevel == "" {
		logLevel = "debug"
	}

	return Config{
		DSN:         dsn,
		Port:        port,
		CORSOrigins: corsOrigins,
		Environment: environment,
		LogLevel:    logLevel,
	}
}

func Run(ctx context.Context, cfg Config) error {
	app, err := builders.NewServerBuilder().
		WithDSN(cfg.DSN).
		WithPort(cfg.Port).
		WithCORSOrigins(cfg.CORSOrigins).
		WithEnvironment(cfg.Environment).
		WithLogLevel(cfg.LogLevel).
		Build(ctx)
	if err != nil {
		return err
	}
	defer app.Close(ctx)

	return app.Start()
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
