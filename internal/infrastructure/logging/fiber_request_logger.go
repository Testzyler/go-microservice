package logging

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func FiberRequestLogger(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		requestID := fmt.Sprint(c.Locals("requestid"))
		if requestID == "<nil>" {
			requestID = ""
		}

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Method()),
			zap.String("path", c.OriginalURL()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", latency),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get("User-Agent")),
		}

		if err != nil {
			logger.Error("http request failed", append(fields, zap.Error(err))...)
			return err
		}

		switch status := c.Response().StatusCode(); {
		case status >= fiber.StatusInternalServerError:
			logger.Error("http request completed", fields...)
		case status >= fiber.StatusBadRequest:
			logger.Warn("http request completed", fields...)
		default:
			logger.Info("http request completed", fields...)
		}

		return nil
	}
}
