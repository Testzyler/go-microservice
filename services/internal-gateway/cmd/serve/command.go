package serve

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/Testzyler/go-microservice/global/pkg/logging"
	"github.com/Testzyler/go-microservice/global/pkg/observability"
	"github.com/Testzyler/go-microservice/services/internal-gateway/internal/config"
	"github.com/Testzyler/go-microservice/services/internal-gateway/internal/proxy"
)

// NewCommand creates the gateway serve command
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve-internal-gateway",
		Short: "Start the internal gateway skeleton",
		Long:  "Start the internal gateway skeleton (auth middleware, health, token issuance).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
}

func run() error {
	ctx := context.Background()

	cfg := config.Load()

	serviceName := viper.GetString("GATEWAY_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "internal-gateway"
	}

	logger, err := logging.New(logging.Config{
		ServiceName: serviceName,
		Environment: cfg.Environment,
		Level:       viper.GetString("LOG_LEVEL"),
		Encoding:    viper.GetString("LOG_ENCODING"),
	})
	if err != nil {
		return err
	}
	defer logger.Sync()

	// Initialize OpenTelemetry
	_, shutdown, err := observability.Init(ctx, observability.Config{
		ServiceName:      serviceName,
		ServiceNamespace: viper.GetString("SERVICE_NAMESPACE"),
		Environment:      cfg.Environment,
		OTLPEndpoint:     viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT"),
		OTLPInsecure:     viper.GetBool("OTEL_EXPORTER_OTLP_INSECURE"),
		Enabled:          cfg.EnableMonitoring,
	})
	if err != nil {
		logger.Error("failed to initialize observability", zap.Error(err))
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(ctx)
	}()

	router, err := proxy.NewRouter(cfg, logger)
	if err != nil {
		logger.Fatal("failed to create router", zap.Error(err))
		return err
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("shutting down gateway...")
		router.Shutdown()
	}()

	return router.Start()
}
