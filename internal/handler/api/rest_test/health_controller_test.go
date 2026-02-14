package rest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	rest "github.com/sklinkert/go-ddd/internal/handler/api/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type probeCheck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type probeResponse struct {
	Status string                `json:"status"`
	Env    string                `json:"env"`
	Checks map[string]probeCheck `json:"checks,omitempty"`
}

type fakeReadinessPinger struct {
	err error
}

func (f fakeReadinessPinger) Ping(_ context.Context) error {
	return f.err
}

func TestHealthController_Liveness(t *testing.T) {
	app := fiber.New()
	rest.NewHealthController(app, "test", fakeReadinessPinger{}, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body probeResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, decodeErr)
	assert.Equal(t, "ok", body.Status)
	assert.Equal(t, "test", body.Env)
}

func TestHealthController_ReadinessReady(t *testing.T) {
	app := fiber.New()
	rest.NewHealthController(app, "test", fakeReadinessPinger{}, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/readyz", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body probeResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, decodeErr)
	assert.Equal(t, "ready", body.Status)
	assert.Equal(t, "up", body.Checks["database"].Status)
}

func TestHealthController_ReadinessNotReady(t *testing.T) {
	app := fiber.New()
	rest.NewHealthController(app, "test", fakeReadinessPinger{
		err: errors.New("db unavailable"),
	}, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/readyz", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	var body probeResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, decodeErr)
	assert.Equal(t, "not_ready", body.Status)
	assert.Equal(t, "down", body.Checks["database"].Status)
	assert.Equal(t, "database ping failed", body.Checks["database"].Message)
}
