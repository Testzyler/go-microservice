package rest

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sklinkert/go-ddd/internal/handler/api/rest/httpx"
	"go.uber.org/zap"
)

const defaultReadinessTimeout = 2 * time.Second

type readinessPinger interface {
	Ping(ctx context.Context) error
}

type HealthController struct {
	environment      string
	dbPinger         readinessPinger
	readinessTimeout time.Duration
	logger           *zap.Logger
}

type dependencyCheck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type readinessResponse struct {
	Status string                     `json:"status"`
	Env    string                     `json:"env"`
	Checks map[string]dependencyCheck `json:"checks"`
}

func NewHealthController(app fiber.Router, environment string, dbPinger readinessPinger, logger *zap.Logger) *HealthController {
	controller := &HealthController{
		environment:      environment,
		dbPinger:         dbPinger,
		readinessTimeout: defaultReadinessTimeout,
		logger:           logger,
	}

	app.Get("/api/v1/healthz", controller.LivenessCheckController)
	app.Get("/api/v1/readyz", controller.ReadinessCheckController)

	return controller
}

func (hc *HealthController) LivenessCheckController(c *fiber.Ctx) error {
	return httpx.OK(c, fiber.Map{
		"status": "ok",
		"env":    hc.environment,
	})
}

func (hc *HealthController) ReadinessCheckController(c *fiber.Ctx) error {
	start := time.Now()

	response := readinessResponse{
		Status: "ready",
		Env:    hc.environment,
		Checks: map[string]dependencyCheck{
			"database": {
				Status: "up",
			},
		},
	}

	if hc.dbPinger == nil {
		response.Status = "not_ready"
		response.Checks["database"] = dependencyCheck{
			Status:  "down",
			Message: "database probe is not configured",
		}
		if hc.logger != nil {
			hc.logger.Warn("readiness check failed: database probe is not configured")
		}
		return c.Status(fiber.StatusServiceUnavailable).JSON(response)
	}

	ctx, cancel := context.WithTimeout(context.Background(), hc.readinessTimeout)
	defer cancel()

	if err := hc.dbPinger.Ping(ctx); err != nil {
		response.Status = "not_ready"
		response.Checks["database"] = dependencyCheck{
			Status:  "down",
			Message: "database ping failed",
		}
		if hc.logger != nil {
			hc.logger.Warn("readiness check failed: database ping",
				zap.Error(err),
				zap.Duration("latency", time.Since(start)),
			)
		}
		return c.Status(fiber.StatusServiceUnavailable).JSON(response)
	}

	if hc.logger != nil {
		hc.logger.Debug("readiness check database ping succeeded",
			zap.Duration("latency", time.Since(start)),
		)
	}

	return httpx.OK(c, response)
}
