package logging

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Environment string
	Level       string
}

func NewLogger(cfg Config) (*zap.Logger, error) {
	environment := normalizeEnvironment(cfg.Environment)

	level := defaultLevel(environment)
	if strings.TrimSpace(cfg.Level) != "" {
		if err := level.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(cfg.Level)))); err != nil {
			return nil, err
		}
	}

	var zapConfig zap.Config
	if environment == "production" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return zapConfig.Build()
}

func normalizeEnvironment(env string) string {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "prod":
		return "production"
	case "dev":
		return "development"
	default:
		if strings.TrimSpace(env) == "" {
			return "development"
		}
		return strings.ToLower(strings.TrimSpace(env))
	}
}

func defaultLevel(environment string) zapcore.Level {
	if environment == "production" {
		return zapcore.InfoLevel
	}
	return zapcore.DebugLevel
}
