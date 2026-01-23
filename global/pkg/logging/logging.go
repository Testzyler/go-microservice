package logging

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	ServiceName string
	Environment string
	Level       string
	Encoding    string
}

func New(cfg Config) (*zap.Logger, error) {
	var zapCfg zap.Config
	devEnv := !strings.EqualFold(cfg.Environment, "prod")
	if devEnv {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}
	if cfg.Encoding != "" {
		zapCfg.Encoding = strings.ToLower(cfg.Encoding)
	}
	if devEnv {
		if strings.EqualFold(zapCfg.Encoding, "console") {
			zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		} else {
			zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		}
	}
	if cfg.Level != "" {
		lvl := zap.NewAtomicLevel()
		if err := lvl.UnmarshalText([]byte(cfg.Level)); err != nil {
			return nil, err
		}
		zapCfg.Level = lvl
	}
	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}
	return logger.With(zap.String("service", cfg.ServiceName)), nil
}
