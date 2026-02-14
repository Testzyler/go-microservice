package rest_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sklinkert/go-ddd/internal/application/command"
	"github.com/sklinkert/go-ddd/internal/application/common"
	"github.com/sklinkert/go-ddd/internal/domain/entities"
	rest "github.com/sklinkert/go-ddd/internal/handler/api/rest"
	"github.com/sklinkert/go-ddd/internal/handler/api/rest/dto/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateProduct(t *testing.T) {
	app := fiber.New()
	mockService := new(MockProductService)
	rest.NewProductController(app, mockService)

	reqBody := map[string]interface{}{
		"name":     "TestProduct",
		"price":    9.99,
		"sellerId": "123e4567-e89b-12d3-a456-426614174000",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	createProductCommandResult := &command.CreateProductCommandResult{
		Result: &common.ProductResult{
			Id:    uuid.New(),
			Name:  "TestProduct",
			Price: 9.99,
		},
	}
	mockService.On("CreateProduct", mock.Anything).Return(createProductCommandResult, nil)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var received response.ProductResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&received)
	assert.NoError(t, decodeErr)
	assert.Equal(t, "TestProduct", received.Name)
	assert.Equal(t, 9.99, received.Price)
	mockService.AssertExpectations(t)
}

func TestGetAllProducts(t *testing.T) {
	app := fiber.New()
	mockService := new(MockProductService)
	rest.NewProductController(app, mockService)

	expectedProducts := []*entities.Product{
		{
			Id:    uuid.New(),
			Name:  "TestProduct1",
			Price: 9.99,
		}, {
			Id:    uuid.New(),
			Name:  "TestProduct2",
			Price: 14.99,
		},
	}

	mockService.On("FindAllProducts").Return(expectedProducts, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var received response.ListProductsResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&received)
	assert.NoError(t, decodeErr)
	assert.Len(t, received.Products, 2)
}
