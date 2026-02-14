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

const traceLevel = "trace"

func NewLogger(cfg Config) (*zap.Logger, error) {
	environment := normalizeEnvironment(cfg.Environment)

	level := defaultLevel(environment)
	rawLevel := strings.ToLower(strings.TrimSpace(cfg.Level))
	if rawLevel != "" {
		if rawLevel == traceLevel {
			level = zapcore.DebugLevel
		} else if err := level.UnmarshalText([]byte(rawLevel)); err != nil {
			return nil, err
		}
	}

	zapConfig := buildConfig(environment)
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	return zapConfig.Build(zap.AddStacktrace(zapcore.ErrorLevel))
}

func IsTraceLevel(level string) bool {
	return strings.EqualFold(strings.TrimSpace(level), traceLevel)
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
