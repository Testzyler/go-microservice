package query

import (
	"github.com/google/uuid"
)

type GetSellerByIdQuery struct {
	Id uuid.UUID
}

func NewGetSellerByIdQuery(id uuid.UUID) *GetSellerByIdQuery {
	return &GetSellerByIdQuery{Id: id}
}

type GetSellerByIdQueryResult = SellerQueryResult
