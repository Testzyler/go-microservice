package mailer

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

	"github.com/Testzyler/go-microservice/global/pkg/async"
	"github.com/Testzyler/go-microservice/global/pkg/config"
	"github.com/Testzyler/go-microservice/global/pkg/email"
	"github.com/Testzyler/go-microservice/global/pkg/logging"
	"github.com/Testzyler/go-microservice/global/pkg/observability"
	"github.com/Testzyler/go-microservice/services/auth/internal/adapters/queue"
	authconfig "github.com/Testzyler/go-microservice/services/auth/internal/config"
)

func NewCommand() *cobra.Command {
	var envFiles string

	cmd := &cobra.Command{
		Use:   "mailer",
		Short: "Run auth mailer worker",
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
		ServiceName: cfg.ServiceName + "-worker-mailer",
		Environment: cfg.Environment,
		Level:       cfg.LogLevel,
		Encoding:    cfg.LogEncoding,
	})
	if err != nil {
		return err
	}
	defer logger.Sync()

	metricsHandler, shutdown, err := observability.Init(ctx, observability.Config{
		ServiceName:      cfg.ServiceName + "-worker-mailer",
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
			queue.QueueCritical: cfg.QueueCritical,
			queue.QueueDefault:  cfg.QueueDefault,
		},
	}, logger)

	var sender email.Sender = email.NoopSender{}
	if cfg.EmailEnabled {
		client, err := email.NewPostmarkClient(email.PostmarkConfig{
			ServerToken:   cfg.PostmarkServerToken,
			From:          cfg.EmailFrom,
			MessageStream: cfg.PostmarkMessageStream,
			Endpoint:      cfg.PostmarkEndpoint,
			Timeout:       10 * time.Second,
		})
		if err != nil {
			return err
		}
		sender = client
	} else {
		logger.Info("email sending disabled")
	}

	handler := queue.NewMailerHandler(logger, sender, cfg.ServiceName)
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
		logger.Error("mailer worker error", zap.Error(err))
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
