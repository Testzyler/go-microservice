package observability

import (
	"context"
	"net/http"

	promclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

type Config struct {
	ServiceName      string
	ServiceNamespace string
	Environment      string
	OTLPEndpoint     string
	OTLPInsecure     bool
	Enabled          bool
}

func Init(ctx context.Context, cfg Config) (http.Handler, func(context.Context) error, error) {
	if !cfg.Enabled {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("monitoring disabled"))
		}), func(context.Context) error { return nil }, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("service.namespace", cfg.ServiceNamespace),
			attribute.String("deployment.environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	exporterOpts := []otlptracegrpc.Option{}
	if cfg.OTLPEndpoint != "" {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint))
	}
	if cfg.OTLPInsecure {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithInsecure())
	}
	traceExporter, err := otlptracegrpc.New(ctx, exporterOpts...)
	if err != nil {
		return nil, nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithBatcher(traceExporter),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	registry := promclient.NewRegistry()
	promExporter, err := prometheus.New(prometheus.WithRegisterer(registry))
	if err != nil {
		return nil, nil, err
	}
	metricsHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(promExporter),
	)
	otel.SetMeterProvider(meterProvider)

	shutdown := func(ctx context.Context) error {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			return err
		}
		return meterProvider.Shutdown(ctx)
	}

	return metricsHandler, shutdown, nil
}
