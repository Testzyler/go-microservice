package queue

import (
	"context"

	"github.com/Testzyler/go-microservice/global/pkg/async"
)

type Publisher struct {
	client *async.Publisher
}

func NewPublisher(client *async.Publisher) *Publisher {
	return &Publisher{client: client}
}

func (p *Publisher) EnqueueWelcomeEmail(ctx context.Context, userID, email string) error {
	task, opts, err := NewWelcomeEmailTask(userID, email)
	if err != nil {
		return err
	}
	_, err = p.client.Enqueue(ctx, task, opts...)
	return err
}

func (p *Publisher) EnqueueSendMFA(ctx context.Context, payload SendMFAPayload) error {
	task, opts, err := NewSendMFATask(payload)
	if err != nil {
		return err
	}
	_, err = p.client.Enqueue(ctx, task, opts...)
	return err
}

func (p *Publisher) EnqueueAuditLog(ctx context.Context, payload AuditLogPayload) error {
	task, opts, err := NewAuditLogTask(payload)
	if err != nil {
		return err
	}
	_, err = p.client.Enqueue(ctx, task, opts...)
	return err
}

func (p *Publisher) EnqueueCleanupSessions(ctx context.Context, graceMinutes int) error {
	task, opts, err := NewCleanupSessionsTask(graceMinutes)
	if err != nil {
		return err
	}
	_, err = p.client.Enqueue(ctx, task, opts...)
	return err
}
