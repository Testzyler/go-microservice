package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/errorx"
)

func TestWriteError_NotFound(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return WriteError(c, errorx.NotFound(errorx.CodeSellerNotFound, errorx.MessageSellerNotFound))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404 but got %d", resp.StatusCode)
	}

	var body ErrorResponse
	if unmarshalErr := json.NewDecoder(resp.Body).Decode(&body); unmarshalErr != nil {
		t.Fatalf("failed to unmarshal response: %v", unmarshalErr)
	}
	if body.Error.Code != errorx.CodeSellerNotFound {
		t.Fatalf("expected code %s but got %s", errorx.CodeSellerNotFound, body.Error.Code)
	}
}

func TestWriteError_InternalFallback(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return WriteError(c, assertErr("unexpected"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500 but got %d", resp.StatusCode)
	}
}

type assertErr string

func (e assertErr) Error() string {
	return string(e)
}
