package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sklinkert/go-ddd/internal/application/command"
	restpkg "github.com/sklinkert/go-ddd/internal/handler/api/rest"
	"github.com/sklinkert/go-ddd/internal/handler/api/rest/dto/request"
	"github.com/sklinkert/go-ddd/internal/handler/api/rest/dto/response"
	"github.com/stretchr/testify/assert"
)

func TestCreateSeller(t *testing.T) {
	mockService := NewMockSellerService()
	app := fiber.New()
	restpkg.NewSellerController(app, mockService)

	sellerJSON, _ := json.Marshal(map[string]string{"name": "TestSeller"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sellers", bytes.NewReader(sellerJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdSeller response.SellerResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&createdSeller)
	assert.NoError(t, decodeErr)
	assert.Equal(t, "TestSeller", createdSeller.Name)
}

func TestPutSeller(t *testing.T) {
	mockService := NewMockSellerService()
	app := fiber.New()
	restpkg.NewSellerController(app, mockService)

	createdSeller, err := mockService.CreateSeller(&command.CreateSellerCommand{Name: "TestSeller"})
	assert.NoError(t, err)

	updateRequest := request.UpdateSellerRequest{
		Id:   createdSeller.Result.Id,
		Name: "updatedName",
	}
	sellerJSON, _ := json.Marshal(updateRequest)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/sellers", bytes.NewReader(sellerJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, reqErr := app.Test(req, -1)
	assert.NoError(t, reqErr)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var receivedResponse response.SellerResponse
	err = json.NewDecoder(resp.Body).Decode(&receivedResponse)
	assert.NoError(t, err)
	assert.Equal(t, updateRequest.Name, receivedResponse.Name)
}

func TestDeleteSeller(t *testing.T) {
	mockService := NewMockSellerService()
	app := fiber.New()
	restpkg.NewSellerController(app, mockService)

	createdSeller, err := mockService.CreateSeller(&command.CreateSellerCommand{Name: "TestSeller"})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/sellers/%s", createdSeller.Result.Id), nil)
	resp, reqErr := app.Test(req, -1)
	assert.NoError(t, reqErr)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestGetSellerById(t *testing.T) {
	mockService := NewMockSellerService()
	app := fiber.New()
	restpkg.NewSellerController(app, mockService)

	createdSeller, err := mockService.CreateSeller(&command.CreateSellerCommand{Name: "TestSeller"})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/sellers/%s", createdSeller.Result.Id), nil)
	resp, reqErr := app.Test(req, -1)
	assert.NoError(t, reqErr)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var fetchedSeller response.SellerResponse
	err = json.NewDecoder(resp.Body).Decode(&fetchedSeller)
	assert.NoError(t, err)
	assert.Equal(t, createdSeller.Result.Id.String(), fetchedSeller.Id)
	assert.Equal(t, createdSeller.Result.Name, fetchedSeller.Name)
}

func TestGetAllSellers(t *testing.T) {
	mockService := NewMockSellerService()
	app := fiber.New()
	restpkg.NewSellerController(app, mockService)

	_, err := mockService.CreateSeller(&command.CreateSellerCommand{Name: "TestSeller1"})
	assert.NoError(t, err)
	_, err = mockService.CreateSeller(&command.CreateSellerCommand{Name: "TestSeller2"})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sellers", nil)
	resp, reqErr := app.Test(req, -1)
	assert.NoError(t, reqErr)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var sellers response.ListSellersResponse
	err = json.NewDecoder(resp.Body).Decode(&sellers)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(sellers.Sellers))
}
