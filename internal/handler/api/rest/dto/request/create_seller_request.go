package request

import (
	"github.com/sklinkert/go-ddd/internal/application/command"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/errorx"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/validate"
)

type CreateSellerRequest struct {
	IdempotencyKey string `json:"idempotencyKey"`
	Name           string `json:"name"`
}

func (req *CreateSellerRequest) Validate() error {
	details := make([]errorx.Detail, 0, 1)
	validate.RequiredString(&details, "name", req.Name, "name is required")
	validate.MinStringLength(&details, "name", req.Name, 2, "name must be at least 2 characters")
	return validate.NewValidationError(details)
}

func (req *CreateSellerRequest) ToCreateSellerCommand() (*command.CreateSellerCommand, error) {
	return &command.CreateSellerCommand{
		IdempotencyKey: req.IdempotencyKey,
		Name:           req.Name,
	}, nil
}
