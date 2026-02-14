package repositories

import (
	"context"

	"github.com/google/uuid"
	entities "github.com/sklinkert/go-ddd/internal/domain/entities"
)

type IdempotencyRepository interface {
	FindByKey(ctx context.Context, key string) (*entities.IdempotencyRecord, error)
	Create(ctx context.Context, record *entities.IdempotencyRecord) (*entities.IdempotencyRecord, error)
	Update(ctx context.Context, record *entities.IdempotencyRecord) (*entities.IdempotencyRecord, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
