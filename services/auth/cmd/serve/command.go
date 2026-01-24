package serve

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/Testzyler/go-microservice/pkg/async"
	"github.com/Testzyler/go-microservice/pkg/config"
	"github.com/Testzyler/go-microservice/pkg/logging"
	"github.com/Testzyler/go-microservice/pkg/observability"
	"github.com/Testzyler/go-microservice/pkg/postgres"
	"github.com/Testzyler/go-microservice/pkg/token"
	"github.com/Testzyler/go-microservice/services/auth/internal/adapters/connect"
	authrepo "github.com/Testzyler/go-microservice/services/auth/internal/adapters/postgres"
	authqueue "github.com/Testzyler/go-microservice/services/auth/internal/adapters/queue"
	"github.com/Testzyler/go-microservice/services/auth/internal/application"
	authconfig "github.com/Testzyler/go-microservice/services/auth/internal/config"
)

func NewCommand() *cobra.Command {
	var envFiles string

	cmd := &cobra.Command{
		Use:   "serve-auth",
		Short: "Start the auth HTTP service",
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

	cfg, err := authconfig.LoadServer(config.ParseEnvFiles(envFiles))
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

	queueClient := async.NewClient(async.RedisConfig{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer queueClient.Close()

	taskPublisher := authqueue.NewPublisher(async.NewPublisher(queueClient, logger))

	tokenSvc := token.NewJWTService(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTokenTTL)
	db, err := postgres.Open(ctx, postgres.Config{
		DSN:             cfg.DatabaseURL,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxIdleTime: cfg.DBConnMaxIdle,
		ConnMaxLifetime: cfg.DBConnMaxLife,
		EnableTracing:   cfg.EnableMonitoring,
	})
	if err != nil {
		return err
	}
	sqlDB, _ := db.DB()
	defer func() {
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}()

	repo, err := authrepo.NewGormUserRepository(db)
	if err != nil {
		return err
	}
	app := application.NewAuthService(repo, tokenSvc, cfg.AccessTokenTTL, taskPublisher, logger)
	server, err := connect.NewServer(app, logger, metricsHandler)
	if err != nil {
		return err
	}

	go func() {
		addr := fmt.Sprintf(":%d", cfg.HTTPPort)
		if err := server.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Error("auth server failed", zap.Error(err))
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("auth server shutting down")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Stop(ctxShutdown)

	return nil
}
