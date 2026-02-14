package middleware

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const (
	ansiReset   = "\033[0m"
	ansiDim     = "\033[2m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiBlue    = "\033[34m"
	ansiMagenta = "\033[35m"
	ansiCyan    = "\033[36m"
)

func RequestLogger(logger *zap.Logger, environment string) fiber.Handler {
	colorEnabled := shouldColorize(environment)

	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)
		status := c.Response().StatusCode()

		requestID := fmt.Sprint(c.Locals("requestid"))
		if requestID == "<nil>" {
			requestID = ""
		}

		summary := formatRequestSummary(colorEnabled, c.Method(), c.OriginalURL(), status, latency)

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Method()),
			zap.String("path", c.OriginalURL()),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get("User-Agent")),
		}

		if err != nil {
			logger.Error(summary, append(fields, zap.Error(err))...)
			return err
		}

		switch {
		case status >= fiber.StatusInternalServerError:
			logger.Error(summary, fields...)
		case status >= fiber.StatusBadRequest:
			logger.Warn(summary, fields...)
		default:
			logger.Info(summary, fields...)
		}

		return nil
	}
}

func shouldColorize(environment string) bool {
	if normalizeEnvironment(environment) == "production" {
		return false
	}
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		return false
	}
	return strings.ToLower(strings.TrimSpace(os.Getenv("TERM"))) != "dumb"
}

func formatRequestSummary(colorEnabled bool, method, path string, status int, latency time.Duration) string {
	methodText := colorMethod(colorEnabled, method)
	pathText := withColor(colorEnabled, ansiCyan, path)
	statusText := colorStatus(colorEnabled, status)
	latencyText := withColor(colorEnabled, ansiDim, latency.String())

	return fmt.Sprintf("%s %s status_code %s %s", methodText, pathText, statusText, latencyText)
}

func colorMethod(enabled bool, method string) string {
	switch method {
	case fiber.MethodGet:
		return withColor(enabled, ansiBlue, method)
	case fiber.MethodPost:
		return withColor(enabled, ansiGreen, method)
	case fiber.MethodPut, fiber.MethodPatch:
		return withColor(enabled, ansiYellow, method)
	case fiber.MethodDelete:
		return withColor(enabled, ansiRed, method)
	default:
		return withColor(enabled, ansiMagenta, method)
	}
}

func colorStatus(enabled bool, status int) string {
	switch {
	case status >= fiber.StatusInternalServerError:
		return withColor(enabled, ansiRed, fmt.Sprintf("%d", status))
	case status >= fiber.StatusBadRequest:
		return withColor(enabled, ansiYellow, fmt.Sprintf("%d", status))
	default:
		return withColor(enabled, ansiGreen, fmt.Sprintf("%d", status))
	}
}

func withColor(enabled bool, color, value string) string {
	if !enabled {
		return value
	}
	return color + value + ansiReset
}
