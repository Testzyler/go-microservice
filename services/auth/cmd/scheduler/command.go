package scheduler

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
		Use:   "serve-auth-scheduler",
		Short: "Run auth scheduler",
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

	cfg, err := authconfig.LoadScheduler(config.ParseEnvFiles(envFiles))
	if err != nil {
		return err
	}

	logger, err := logging.New(logging.Config{
		ServiceName: cfg.ServiceName,
		Environment: cfg.Environment,
		Level:       cfg.LogLevel,
		Encoding:    cfg.LogEncoding,
	})
	if err != nil {
		return err
	}
	defer logger.Sync()

	metricsHandler, shutdown, err := observability.Init(ctx, observability.Config{
		ServiceName:      cfg.ServiceName,
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

	scheduler := async.NewScheduler(async.RedisConfig{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}, logger)

	if _, err := registerCleanup(scheduler, cfg.CleanupCron, cfg.CleanupGraceMins); err != nil {
		return err
	}
	if _, err := registerHeartbeat(scheduler, cfg.HeartbeatCron, cfg.HeartbeatQueue); err != nil {
		return err
	}

	metricsServer := startMetricsServer(cfg.Metrics, metricsHandler, logger)

	logger.Info("scheduler starting",
		zap.String("cleanup_cron", cfg.CleanupCron),
		zap.String("heartbeat_cron", cfg.HeartbeatCron),
		zap.String("redis_addr", cfg.RedisAddr),
	)

	go func() {
		if err := scheduler.Run(); err != nil && err != asynq.ErrServerClosed {
			logger.Error("scheduler error", zap.Error(err))
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("scheduler shutting down")
	scheduler.Shutdown()

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = metricsServer.Shutdown(ctxShutdown)
	return nil
}

func registerCleanup(s *asynq.Scheduler, cron string, graceMinutes int) (string, error) {
	task, opts, err := queue.NewCleanupSessionsTask(queue.CleanupSessionsPayload{
		GracePeriodMinutes: graceMinutes,
	})
	if err != nil {
		return "", err
	}
	return s.Register(cron, task, opts...)
}

func registerHeartbeat(s *asynq.Scheduler, cron string, queueName string) (string, error) {
	task, opts, err := queue.NewHeartbeatTask(queue.HeartbeatPayload{
		Message: "auth scheduler heartbeat",
	})
	if err != nil {
		return "", err
	}
	opts = append(opts, asynq.Queue(queueName))
	return s.Register(cron, task, opts...)
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
