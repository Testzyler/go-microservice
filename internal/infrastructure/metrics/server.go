package metrics

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const defaultMetricsPort = ":9090"

type Server struct {
	server *http.Server
	logger *zap.Logger
	port   string
}

func NewServer(port string, logger *zap.Logger) *Server {
	normalizedPort := strings.TrimSpace(port)
	if normalizedPort == "" {
		normalizedPort = defaultMetricsPort
	}
	if !strings.HasPrefix(normalizedPort, ":") {
		normalizedPort = ":" + normalizedPort
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &Server{
		server: &http.Server{
			Addr:              normalizedPort,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		logger: logger,
		port:   normalizedPort,
	}
}

func (s *Server) Start() error {
	if s == nil || s.server == nil {
		return nil
	}
	if s.logger != nil {
		s.logger.Info("starting metrics server", zap.String("port", s.port))
	}
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil || s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

