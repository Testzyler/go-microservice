package request

import (
	"github.com/google/uuid"
	"github.com/sklinkert/go-ddd/internal/application/command"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/errorx"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/validate"
)

type CreateProductRequest struct {
	IdempotencyKey string  `json:"idempotencyKey"`
	Name           string  `json:"name"`
	Price          float64 `json:"price"`
	SellerId       string  `json:"sellerId"`
}

func (req *CreateProductRequest) Validate() error {
	details := make([]errorx.Detail, 0, 3)
	validate.RequiredString(&details, "name", req.Name, "name is required")
	validate.MinStringLength(&details, "name", req.Name, 2, "name must be at least 2 characters")
	validate.PositiveFloat(&details, "price", req.Price, "price must be greater than 0")
	validate.UUIDString(&details, "sellerId", req.SellerId, "sellerId must be a valid UUID")
	return validate.NewValidationError(details)
}

func (req *CreateProductRequest) ToCreateProductCommand() (*command.CreateProductCommand, error) {
	sellerId, err := uuid.Parse(req.SellerId)
	if err != nil {
		return nil, errorx.InvalidArgument(errorx.CodeProductInvalidSellerID, errorx.MessageProductInvalidSellerID)
	}

	return &command.CreateProductCommand{
		IdempotencyKey: req.IdempotencyKey,
		Name:           req.Name,
		Price:          req.Price,
		SellerId:       sellerId,
	}, nil
}
