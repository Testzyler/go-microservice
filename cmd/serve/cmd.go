package serve

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sklinkert/go-ddd/internal/builders"
	"github.com/sklinkert/go-ddd/internal/infrastructure/metrics"
)

type Config struct {
	DSN            string
	Port           string
	CORSOrigins    []string
	Environment    string
	LogLevel       string
	MetricsEnabled bool
	MetricsPort    string
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

	metricsEnabled := true
	if raw := strings.TrimSpace(os.Getenv("METRICS_ENABLED")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err == nil {
			metricsEnabled = parsed
		}
	}

	metricsPort := strings.TrimSpace(os.Getenv("METRICS_PORT"))
	if metricsPort == "" {
		metricsPort = ":9090"
	}

	return Config{
		DSN:            dsn,
		Port:           port,
		CORSOrigins:    corsOrigins,
		Environment:    environment,
		LogLevel:       logLevel,
		MetricsEnabled: metricsEnabled,
		MetricsPort:    metricsPort,
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

	var metricsServer *metrics.Server
	var metricsErrCh <-chan error
	if cfg.MetricsEnabled {
		metricsServer = metrics.NewServer(cfg.MetricsPort, nil)
		ch := make(chan error, 1)
		metricsErrCh = ch
		go func() {
			ch <- metricsServer.Start()
		}()
	}

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- app.Start()
	}()

	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case serveErr := <-serverErrCh:
		metricsShutdownErr := shutdownMetricsServer(context.Background(), metricsServer)
		metricsServeErr := awaitOptionalError(metricsErrCh)
		closeErr := app.Close(context.Background())
		if isExpectedShutdownError(metricsServeErr) {
			return errors.Join(serveErr, metricsShutdownErr, closeErr)
		}
		return errors.Join(serveErr, metricsShutdownErr, closeErr, metricsServeErr)
	case metricsServeErr := <-metricsErrCh:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownErr := app.Shutdown(shutdownCtx)
		closeErr := app.Close(shutdownCtx)
		serveErr := <-serverErrCh
		if isExpectedShutdownError(serveErr) {
			return errors.Join(metricsServeErr, shutdownErr, closeErr)
		}
		return errors.Join(metricsServeErr, shutdownErr, closeErr, serveErr)
	case <-signalCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		metricsShutdownErr := shutdownMetricsServer(shutdownCtx, metricsServer)
		shutdownErr := app.Shutdown(shutdownCtx)
		closeErr := app.Close(shutdownCtx)
		serveErr := <-serverErrCh
		metricsServeErr := awaitOptionalError(metricsErrCh)
		if isExpectedShutdownError(serveErr) && isExpectedShutdownError(metricsServeErr) {
			return errors.Join(metricsShutdownErr, shutdownErr, closeErr)
		}
		return errors.Join(metricsShutdownErr, shutdownErr, closeErr, serveErr, metricsServeErr)
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

func shutdownMetricsServer(ctx context.Context, server *metrics.Server) error {
	if server == nil {
		return nil
	}
	return server.Shutdown(ctx)
}

func awaitOptionalError(ch <-chan error) error {
	if ch == nil {
		return nil
	}
	return <-ch
}
