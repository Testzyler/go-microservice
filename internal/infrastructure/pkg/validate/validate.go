package validate

import (
	"strings"

	"github.com/google/uuid"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/errorx"
)

func RequiredString(details *[]errorx.Detail, field, value, message string) {
	if strings.TrimSpace(value) == "" {
		*details = append(*details, errorx.Detail{
			Field:   field,
			Message: message,
		})
	}
}

func MinStringLength(details *[]errorx.Detail, field, value string, min int, message string) {
	if len(strings.TrimSpace(value)) < min {
		*details = append(*details, errorx.Detail{
			Field:   field,
			Message: message,
		})
	}
}

func PositiveFloat(details *[]errorx.Detail, field string, value float64, message string) {
	if value <= 0 {
		*details = append(*details, errorx.Detail{
			Field:   field,
			Message: message,
		})
	}
}

func UUIDString(details *[]errorx.Detail, field, value, message string) {
	if _, err := uuid.Parse(value); err != nil {
		*details = append(*details, errorx.Detail{
			Field:   field,
			Message: message,
		})
	}
}

func RequiredUUID(details *[]errorx.Detail, field string, value uuid.UUID, message string) {
	if value == uuid.Nil {
		*details = append(*details, errorx.Detail{
			Field:   field,
			Message: message,
		})
	}
}

func NewValidationError(details []errorx.Detail) error {
	if len(details) == 0 {
		return nil
	}
	return errorx.InvalidArgument(errorx.CodeValidationError, errorx.MessageValidationErr, details...)
}
