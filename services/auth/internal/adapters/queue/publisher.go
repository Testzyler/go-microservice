package queue

import (
	"context"

	"github.com/Testzyler/go-microservice/pkg/async"
)

type Publisher struct {
	client *async.Publisher
}

func NewPublisher(client *async.Publisher) *Publisher {
	return &Publisher{client: client}
}

func (p *Publisher) EnqueueWelcomeEmail(ctx context.Context, userID, email string) error {
	payload := WelcomeEmailPayload{
		UserID: userID,
		Email:  email,
	}
	async.InjectTrace(ctx, &payload.Trace)
	task, opts, err := NewWelcomeEmailTask(payload)
	if err != nil {
		return err
	}
	_, err = p.client.Enqueue(ctx, task, opts...)
	return err
}

func (p *Publisher) EnqueueSendMFA(ctx context.Context, payload SendMFAPayload) error {
	async.InjectTrace(ctx, &payload.Trace)
	task, opts, err := NewSendMFATask(payload)
	if err != nil {
		return err
	}
	_, err = p.client.Enqueue(ctx, task, opts...)
	return err
}

func (p *Publisher) EnqueueAuditLog(ctx context.Context, payload AuditLogPayload) error {
	async.InjectTrace(ctx, &payload.Trace)
	task, opts, err := NewAuditLogTask(payload)
	if err != nil {
		return err
	}
	_, err = p.client.Enqueue(ctx, task, opts...)
	return err
}

func (p *Publisher) EnqueueCleanupSessions(ctx context.Context, graceMinutes int) error {
	payload := CleanupSessionsPayload{GracePeriodMinutes: graceMinutes}
	async.InjectTrace(ctx, &payload.Trace)
	task, opts, err := NewCleanupSessionsTask(payload)
	if err != nil {
		return err
	}
	_, err = p.client.Enqueue(ctx, task, opts...)
	return err
}
