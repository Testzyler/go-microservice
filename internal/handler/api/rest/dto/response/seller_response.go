package response

import "time"

type SellerResponse struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ListSellersResponse struct {
	Sellers []*SellerResponse `json:"sellers"`
}
