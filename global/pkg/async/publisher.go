package async

import (
	"context"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Publisher struct {
	client *asynq.Client
	tracer trace.Tracer
	logger *zap.Logger
}

func NewPublisher(client *asynq.Client, logger *zap.Logger) *Publisher {
	tracer := otel.Tracer("async.publisher")
	return &Publisher{client: client, tracer: tracer, logger: logger}
}

func (p *Publisher) Enqueue(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx, span := p.tracer.Start(ctx, "async.enqueue", trace.WithSpanKind(trace.SpanKindProducer))
	span.SetAttributes(
		attribute.String("task.type", task.Type()),
		attribute.String("messaging.system", "asynq"),
		attribute.String("messaging.operation", "publish"),
		attribute.String("messaging.message_type", task.Type()),
	)
	defer span.End()

	info, err := p.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
		if p.logger != nil {
			p.logger.Error("enqueue task failed", zap.String("type", task.Type()), zap.Error(err))
		}
		return nil, err
	}
	span.SetAttributes(
		attribute.String("messaging.message_id", info.ID),
		attribute.String("messaging.destination", info.Queue),
	)

	if p.logger != nil {
		p.logger.Debug("enqueued task",
			zap.String("id", info.ID),
			zap.String("queue", info.Queue),
			zap.String("type", task.Type()),
		)
	}
	return info, nil
}
