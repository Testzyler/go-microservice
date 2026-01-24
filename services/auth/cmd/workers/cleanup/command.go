package cleanup

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/Testzyler/go-microservice/pkg/async"
	"github.com/Testzyler/go-microservice/pkg/config"
	"github.com/Testzyler/go-microservice/pkg/logging"
	"github.com/Testzyler/go-microservice/pkg/observability"
	"github.com/Testzyler/go-microservice/services/auth/internal/adapters/queue"
	authconfig "github.com/Testzyler/go-microservice/services/auth/internal/config"
)

func NewCommand() *cobra.Command {
	var envFiles string

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Run auth cleanup worker",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(envFiles)
		},
	}
	cmd.Flags().StringVar(&envFiles, "env-files", "", "comma-separated env files to load")
	return cmd
}

func run(envFiles string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := authconfig.LoadWorker(config.ParseEnvFiles(envFiles))
	if err != nil {
		return err
	}

	logger, err := logging.New(logging.Config{
		ServiceName: cfg.ServiceName + "-worker-cleanup",
		Environment: cfg.Environment,
		Level:       cfg.LogLevel,
		Encoding:    cfg.LogEncoding,
	})
	if err != nil {
		return err
	}
	defer logger.Sync()

	metricsHandler, shutdown, err := observability.Init(ctx, observability.Config{
		ServiceName:      cfg.ServiceName + "-worker-cleanup",
		ServiceNamespace: cfg.ServiceNamespace,
		Environment:      cfg.Environment,
		OTLPEndpoint:     cfg.OTLPEndpoint,
		OTLPInsecure:     cfg.OTLPInsecure,
		Enabled:          cfg.EnableMonitoring,
	})
	if err != nil {
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()

	server := async.NewServer(async.RedisConfig{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}, async.ServerConfig{
		Concurrency: cfg.AsynqConcurrency,
		Queues: map[string]int{
			queue.QueueLow:       cfg.QueueLow,
			queue.QueueHeartbeat: cfg.QueueHeartbeat,
		},
	}, logger)

	handler := queue.NewCleanupHandler(logger)
	mux := asynq.NewServeMux()
	handler.Register(mux)

	metricsServer := startMetricsServer(cfg.Metrics, metricsHandler, logger)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Run(mux)
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal")
	case err := <-serverErr:
		logger.Error("cleanup worker error", zap.Error(err))
	}

	server.Shutdown()

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = metricsServer.Shutdown(ctxShutdown)

	return nil
}

func startMetricsServer(port int, handler http.Handler, logger *zap.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", handler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server error", zap.Error(err))
		}
	}()
	return server
}
