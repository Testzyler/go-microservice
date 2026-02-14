package query

import (
	"github.com/google/uuid"
)

type GetProductByIdQuery struct {
	Id uuid.UUID
}

func NewGetProductByIdQuery(id uuid.UUID) *GetProductByIdQuery {
	return &GetProductByIdQuery{Id: id}
}

type GetProductByIdQueryResult = ProductQueryResult
