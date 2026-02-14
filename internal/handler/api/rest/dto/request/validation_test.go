package request

import (
	"testing"

	"github.com/google/uuid"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/errorx"
)

func TestCreateSellerRequest_Validate(t *testing.T) {
	req := &CreateSellerRequest{Name: "A"}

	err := req.Validate()
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if errorx.KindOf(err) != errorx.KindInvalidArgument {
		t.Fatalf("expected invalid argument error kind")
	}
}

func TestCreateProductRequest_Validate(t *testing.T) {
	req := &CreateProductRequest{
		Name:     "",
		Price:    0,
		SellerId: "not-a-uuid",
	}

	err := req.Validate()
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if errorx.KindOf(err) != errorx.KindInvalidArgument {
		t.Fatalf("expected invalid argument error kind")
	}
}

func TestUpdateSellerRequest_Validate(t *testing.T) {
	req := &UpdateSellerRequest{
		Id:   uuid.Nil,
		Name: "",
	}

	err := req.Validate()
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if errorx.KindOf(err) != errorx.KindInvalidArgument {
		t.Fatalf("expected invalid argument error kind")
	}
}
