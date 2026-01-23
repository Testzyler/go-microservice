package async

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type ServerConfig struct {
	Concurrency int
	Queues      map[string]int
}

func NewClient(cfg RedisConfig) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

func NewServer(cfg RedisConfig, serverCfg ServerConfig, logger *zap.Logger) *asynq.Server {
	asynqLogger := &zapAsynqLogger{logger: logger}
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		},
		asynq.Config{
			Concurrency: serverCfg.Concurrency,
			Queues:      serverCfg.Queues,
			Logger:      asynqLogger,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.Error("task failed", zap.String("type", task.Type()), zap.Error(err))
			}),
			RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
				base := math.Min(float64(2*n), 60)
				jitter := rand.Float64() * 0.3 * base
				return time.Duration(base+jitter) * time.Second
			},
		},
	)
}

func NewScheduler(cfg RedisConfig, logger *zap.Logger) *asynq.Scheduler {
	return asynq.NewScheduler(
		asynq.RedisClientOpt{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		},
		&asynq.SchedulerOpts{Logger: &zapAsynqLogger{logger: logger}},
	)
}

type zapAsynqLogger struct {
	logger *zap.Logger
}

func (l *zapAsynqLogger) Debug(args ...interface{}) { l.logger.Sugar().Debug(args...) }
func (l *zapAsynqLogger) Info(args ...interface{})  { l.logger.Sugar().Info(args...) }
func (l *zapAsynqLogger) Warn(args ...interface{})  { l.logger.Sugar().Warn(args...) }
func (l *zapAsynqLogger) Error(args ...interface{}) { l.logger.Sugar().Error(args...) }
func (l *zapAsynqLogger) Fatal(args ...interface{}) { l.logger.Sugar().Fatal(args...) }
