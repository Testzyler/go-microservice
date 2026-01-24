package email

import "context"

type NoopSender struct{}

func (NoopSender) Send(ctx context.Context, msg Message) error {
	return nil
}
