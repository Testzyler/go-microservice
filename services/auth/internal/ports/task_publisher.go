package ports

import "context"

type TaskPublisher interface {
	EnqueueWelcomeEmail(ctx context.Context, userID, email string) error
}
