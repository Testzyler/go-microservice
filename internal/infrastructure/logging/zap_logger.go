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

	zapConfig := buildConfig(environment)
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	return zapConfig.Build(zap.AddStacktrace(zapcore.ErrorLevel))
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

func buildConfig(environment string) zap.Config {
	if environment == "production" {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		return cfg
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.DisableStacktrace = true
	return cfg
}
