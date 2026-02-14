package response

import "time"

type ProductResponse struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ListProductsResponse struct {
	Products []*ProductResponse `json:"products"`
}
