package config

import (
	"time"

	"github.com/spf13/viper"

	envconfig "github.com/Testzyler/go-microservice/pkg/config"
)

// Config holds gateway configuration
type Config struct {
	Port           string `mapstructure:"GATEWAY_PORT"`
	Environment    string `mapstructure:"ENVIRONMENT"`
	AuthServiceURL string `mapstructure:"AUTH_SERVICE_URL"`

	// JWT configuration
	JWTSecret      string        `mapstructure:"JWT_SECRET"`
	JWTIssuer      string        `mapstructure:"JWT_ISSUER"`
	JWTAudience    string        `mapstructure:"JWT_AUDIENCE"`
	AccessTokenTTL time.Duration `mapstructure:"ACCESS_TOKEN_TTL"`

	EnableMonitoring bool `mapstructure:"ENABLE_MONITORING"`

	// Rate limiting
	RateLimitEnabled bool `mapstructure:"RATE_LIMIT_ENABLED"`
	RateLimitRPS     int  `mapstructure:"RATE_LIMIT_RPS"`
	RateLimitBurst   int  `mapstructure:"RATE_LIMIT_BURST"`

	// CORS
	CORSAllowOrigins []string `mapstructure:"CORS_ALLOW_ORIGINS"`
	CORSAllowMethods []string `mapstructure:"CORS_ALLOW_METHODS"`
	CORSAllowHeaders []string `mapstructure:"CORS_ALLOW_HEADERS"`

	// Timeouts
	ReadTimeout  time.Duration `mapstructure:"READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"WRITE_TIMEOUT"`
	IdleTimeout  time.Duration `mapstructure:"IDLE_TIMEOUT"`
}

// Load loads configuration from environment
func Load() *Config {
	viper.AutomaticEnv()

	viper.SetDefault("GATEWAY_PORT", "8080")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("AUTH_SERVICE_URL", "http://localhost:8081")
	viper.SetDefault("JWT_SECRET", "")
	viper.SetDefault("JWT_ISSUER", "go-microservice")
	viper.SetDefault("JWT_AUDIENCE", "go-microservice-users")
	viper.SetDefault("ACCESS_TOKEN_TTL", "15m")
	viper.SetDefault("RATE_LIMIT_ENABLED", true)
	viper.SetDefault("RATE_LIMIT_RPS", 100)
	viper.SetDefault("RATE_LIMIT_BURST", 200)
	viper.SetDefault("CORS_ALLOW_ORIGINS", "*")
	viper.SetDefault("CORS_ALLOW_METHODS", "GET,POST,PUT,DELETE,OPTIONS")
	viper.SetDefault("CORS_ALLOW_HEADERS", "Origin,Content-Type,Accept,Authorization")
	viper.SetDefault("READ_TIMEOUT", "15s")
	viper.SetDefault("WRITE_TIMEOUT", "15s")
	viper.SetDefault("IDLE_TIMEOUT", "60s")
	viper.SetDefault("ENABLE_MONITORING", true)

	cfg := &Config{}
	_ = envconfig.Load(cfg, envconfig.Options{
		Defaults: map[string]interface{}{
			"GATEWAY_PORT":       "8080",
			"ENVIRONMENT":        "development",
			"AUTH_SERVICE_URL":   "http://localhost:8081",
			"JWT_SECRET":         "",
			"JWT_ISSUER":         "go-microservice",
			"JWT_AUDIENCE":       "go-microservice-users",
			"ACCESS_TOKEN_TTL":   "15m",
			"RATE_LIMIT_ENABLED": true,
			"RATE_LIMIT_RPS":     100,
			"RATE_LIMIT_BURST":   200,
			"CORS_ALLOW_ORIGINS": "*",
			"CORS_ALLOW_METHODS": "GET,POST,PUT,DELETE,OPTIONS",
			"CORS_ALLOW_HEADERS": "Origin,Content-Type,Accept,Authorization",
			"READ_TIMEOUT":       "15s",
			"WRITE_TIMEOUT":      "15s",
			"IDLE_TIMEOUT":       "60s",
			"ENABLE_MONITORING":  true,
		},
	})

	if cfg.CORSAllowOrigins == nil {
		cfg.CORSAllowOrigins = splitString(viper.GetString("CORS_ALLOW_ORIGINS"))
	}
	if cfg.CORSAllowMethods == nil {
		cfg.CORSAllowMethods = splitString(viper.GetString("CORS_ALLOW_METHODS"))
	}
	if cfg.CORSAllowHeaders == nil {
		cfg.CORSAllowHeaders = splitString(viper.GetString("CORS_ALLOW_HEADERS"))
	}
	return cfg
}

func splitString(s string) []string {
	if s == "" {
		return nil
	}
	result := []string{}
	current := ""
	for _, c := range s {
		if c == ',' {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
