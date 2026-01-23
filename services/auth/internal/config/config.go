package config

import (
	"github.com/Testzyler/go-microservice/global/pkg/config"
	"time"
)

type Common struct {
	ServiceName      string `mapstructure:"auth_service_name"`
	ServiceNamespace string `mapstructure:"service_namespace"`
	Environment      string `mapstructure:"environment"`
	LogLevel         string `mapstructure:"log_level"`
	LogEncoding      string `mapstructure:"log_encoding"`
	EnableMonitoring bool   `mapstructure:"enable_monitoring"`
	OTLPEndpoint     string `mapstructure:"otel_exporter_otlp_endpoint"`
	OTLPInsecure     bool   `mapstructure:"otel_exporter_otlp_insecure"`
}

type Async struct {
	RedisAddr        string `mapstructure:"auth_async_redis_addr"`
	RedisPassword    string `mapstructure:"auth_async_redis_password"`
	RedisDB          int    `mapstructure:"auth_async_redis_db"`
	QueueCritical    int    `mapstructure:"auth_async_queue_critical"`
	QueueDefault     int    `mapstructure:"auth_async_queue_default"`
	QueueLow         int    `mapstructure:"auth_async_queue_low"`
	QueueAudit       int    `mapstructure:"auth_async_queue_audit"`
	QueueHeartbeat   int    `mapstructure:"auth_async_queue_heartbeat"`
	AsynqConcurrency int    `mapstructure:"auth_async_concurrency"`
}

type ServerConfig struct {
	Common
	HTTPPort       int           `mapstructure:"auth_http_port"`
	Metrics        int           `mapstructure:"auth_metrics_port"`
	JWTSecret      string        `mapstructure:"auth_jwt_secret"`
	JWTIssuer      string        `mapstructure:"auth_jwt_issuer"`
	JWTAudience    string        `mapstructure:"auth_jwt_audience"`
	AccessTokenTTL time.Duration `mapstructure:"auth_access_token_ttl"`
	DatabaseURL    string        `mapstructure:"auth_database_url"`
	DBMaxOpenConns int           `mapstructure:"auth_db_max_open_conns"`
	DBMaxIdleConns int           `mapstructure:"auth_db_max_idle_conns"`
	DBConnMaxIdle  time.Duration `mapstructure:"auth_db_conn_max_idle"`
	DBConnMaxLife  time.Duration `mapstructure:"auth_db_conn_max_life"`
	Async
}

type WorkerConfig struct {
	Common
	Metrics int `mapstructure:"auth_metrics_port"`
	Async
}

type SchedulerConfig struct {
	Common
	Metrics          int    `mapstructure:"auth_scheduler_metrics_port"`
	CleanupCron      string `mapstructure:"auth_scheduler_cleanup_cron"`
	HeartbeatCron    string `mapstructure:"auth_scheduler_heartbeat_cron"`
	HeartbeatQueue   string `mapstructure:"auth_scheduler_heartbeat_queue"`
	CleanupGraceMins int    `mapstructure:"auth_scheduler_cleanup_grace_minutes"`
	Async
}

func LoadServer(envFiles []string) (ServerConfig, error) {
	cfg := ServerConfig{}
	err := config.Load(&cfg, config.Options{
		EnvFiles: envFiles,
		Defaults: defaultMap(),
	})
	return cfg, err
}

func LoadWorker(envFiles []string) (WorkerConfig, error) {
	cfg := WorkerConfig{}
	err := config.Load(&cfg, config.Options{
		EnvFiles: envFiles,
		Defaults: defaultMap(),
	})
	return cfg, err
}

func LoadScheduler(envFiles []string) (SchedulerConfig, error) {
	cfg := SchedulerConfig{}
	err := config.Load(&cfg, config.Options{
		EnvFiles: envFiles,
		Defaults: defaultMap(),
	})
	return cfg, err
}

func defaultMap() map[string]interface{} {
	return map[string]interface{}{
		"auth_service_name":                    "auth",
		"service_namespace":                    "auth",
		"environment":                          "local",
		"log_level":                            "debug",
		"log_encoding":                         "console",
		"enable_monitoring":                    true,
		"otel_exporter_otlp_endpoint":          "",
		"otel_exporter_otlp_insecure":          true,
		"auth_http_port":                       8081,
		"auth_metrics_port":                    9002,
		"auth_async_redis_addr":                "localhost:6379",
		"auth_async_redis_password":            "",
		"auth_async_redis_db":                  0,
		"auth_async_queue_critical":            10,
		"auth_async_queue_default":             10,
		"auth_async_queue_low":                 5,
		"auth_async_queue_audit":               5,
		"auth_async_queue_heartbeat":           2,
		"auth_async_concurrency":               10,
		"auth_scheduler_metrics_port":          9003,
		"auth_scheduler_cleanup_cron":          "@every 30m",
		"auth_scheduler_heartbeat_cron":        "@every 1m",
		"auth_scheduler_heartbeat_queue":       "heartbeat",
		"auth_scheduler_cleanup_grace_minutes": 60,
		"auth_jwt_secret":                      "dev-secret-change-me",
		"auth_jwt_issuer":                      "go-microservice",
		"auth_jwt_audience":                    "go-microservice-users",
		"auth_access_token_ttl":                "15m",
		"auth_database_url":                    "postgres://postgres:postgres@localhost:5432/auth?sslmode=disable",
		"auth_db_max_open_conns":               20,
		"auth_db_max_idle_conns":               10,
		"auth_db_conn_max_idle":                "10m",
		"auth_db_conn_max_life":                "60m",
	}
}
