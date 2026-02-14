package serve

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- app.Start()
	}()

	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case serveErr := <-serverErrCh:
		closeErr := app.Close(context.Background())
		return errors.Join(serveErr, closeErr)
	case <-signalCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownErr := app.Shutdown(shutdownCtx)
		closeErr := app.Close(shutdownCtx)
		serveErr := <-serverErrCh
		if isExpectedShutdownError(serveErr) {
			return errors.Join(shutdownErr, closeErr)
		}
		return errors.Join(shutdownErr, closeErr, serveErr)
	}
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

func isExpectedShutdownError(err error) bool {
	if err == nil {
		return true
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "closed network connection") ||
		strings.Contains(message, "server closed") ||
		strings.Contains(message, "shutdown")
}
