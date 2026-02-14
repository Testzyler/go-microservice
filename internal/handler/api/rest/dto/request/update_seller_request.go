package request

import (
	"github.com/google/uuid"
	"github.com/sklinkert/go-ddd/internal/application/command"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/errorx"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/validate"
)

type UpdateSellerRequest struct {
	IdempotencyKey string    `json:"idempotencyKey"`
	Id             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
}

func (req *UpdateSellerRequest) Validate() error {
	details := make([]errorx.Detail, 0, 2)
	validate.RequiredUUID(&details, "id", req.Id, "id is required")
	validate.RequiredString(&details, "name", req.Name, "name is required")
	validate.MinStringLength(&details, "name", req.Name, 2, "name must be at least 2 characters")
	return validate.NewValidationError(details)
}

func (req *UpdateSellerRequest) ToUpdateSellerCommand() (*command.UpdateSellerCommand, error) {
	return &command.UpdateSellerCommand{
		IdempotencyKey: req.IdempotencyKey,
		Id:             req.Id,
		Name:           req.Name,
	}, nil
}
