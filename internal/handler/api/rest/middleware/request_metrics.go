package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricsRegisterOnce sync.Once

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed by the API server.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of in-flight HTTP requests.",
		},
	)
)

func RequestMetrics() fiber.Handler {
	metricsRegisterOnce.Do(func() {
		prometheus.MustRegister(httpRequestsTotal, httpRequestDurationSeconds, httpRequestsInFlight)
	})

	return func(c *fiber.Ctx) error {
		start := time.Now()
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		err := c.Next()
		status := strconv.Itoa(c.Response().StatusCode())
		path := c.Path()
		if route := c.Route(); route != nil && route.Path != "" {
			path = route.Path
		}

		httpRequestsTotal.WithLabelValues(c.Method(), path, status).Inc()
		httpRequestDurationSeconds.WithLabelValues(c.Method(), path, status).Observe(time.Since(start).Seconds())

		return err
	}
}

