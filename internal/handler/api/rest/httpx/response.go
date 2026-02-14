package httpx

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/errorx"
)

type ErrorBody struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Details []errorx.Detail `json:"details,omitempty"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func OK(c *fiber.Ctx, body any) error {
	return c.Status(fiber.StatusOK).JSON(body)
}

func Created(c *fiber.Ctx, body any) error {
	return c.Status(fiber.StatusCreated).JSON(body)
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func WriteError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	status := fiber.StatusInternalServerError
	message := errorx.MessageInternalServerError

	switch errorx.KindOf(err) {
	case errorx.KindInvalidArgument:
		status = fiber.StatusBadRequest
		message = err.Error()
	case errorx.KindNotFound:
		status = fiber.StatusNotFound
		message = err.Error()
	case errorx.KindConflict:
		status = fiber.StatusConflict
		message = err.Error()
	case errorx.KindUnauthenticated:
		status = fiber.StatusUnauthorized
		message = err.Error()
	case errorx.KindForbidden:
		status = fiber.StatusForbidden
		message = err.Error()
	case errorx.KindUnavailable:
		status = fiber.StatusServiceUnavailable
		message = err.Error()
	default:
		// Preserve explicit error message if caller already wrapped it.
		if strings.TrimSpace(err.Error()) != "" && errorx.Code(err) != errorx.CodeInternal {
			message = err.Error()
		}
	}

	return c.Status(status).JSON(ErrorResponse{
		Error: ErrorBody{
			Code:    errorx.Code(err),
			Message: message,
			Details: errorx.Details(err),
		},
	})
}

type validatable interface {
	Validate() error
}

func BindAndValidate(c *fiber.Ctx, body any) error {
	if err := c.BodyParser(body); err != nil {
		return errorx.InvalidArgument(errorx.CodeHTTPBadRequest, errorx.MessageInvalidRequestBody)
	}

	if v, ok := body.(validatable); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func ParseUUIDParam(c *fiber.Ctx, param, code, message string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(param))
	if err != nil {
		return uuid.Nil, errorx.InvalidArgument(code, message)
	}
	return id, nil
}

func ResolveIdempotencyKey(c *fiber.Ctx, bodyValue string) string {
	headerValue := strings.TrimSpace(c.Get("Idempotency-Key"))
	if headerValue != "" {
		return headerValue
	}
	return strings.TrimSpace(bodyValue)
}
