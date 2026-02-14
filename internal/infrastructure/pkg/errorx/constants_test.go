package errorx

import (
	"regexp"
	"testing"
)

func TestErrorCodes_AreUniqueAndWellFormed(t *testing.T) {
	codes := map[string]string{
		"CodeInternal":               CodeInternal,
		"CodeValidationError":        CodeValidationError,
		"CodeHTTPBadRequest":         CodeHTTPBadRequest,
		"CodeIdempotencyInProgress":  CodeIdempotencyInProgress,
		"CodeIdempotencyKeyReused":   CodeIdempotencyKeyReused,
		"CodeSellerNotFound":         CodeSellerNotFound,
		"CodeSellerInvalidID":        CodeSellerInvalidID,
		"CodeSellerIDRequired":       CodeSellerIDRequired,
		"CodeSellerQueryRequired":    CodeSellerQueryRequired,
		"CodeProductNotFound":        CodeProductNotFound,
		"CodeProductInvalidID":       CodeProductInvalidID,
		"CodeProductIDRequired":      CodeProductIDRequired,
		"CodeProductQueryRequired":   CodeProductQueryRequired,
		"CodeProductInvalidSellerID": CodeProductInvalidSellerID,
	}

	pattern := regexp.MustCompile(`^(SYS|BIZ)-[A-Z]{3,4}-\d{3}-\d{3}$`)
	seen := make(map[string]string, len(codes))

	for name, code := range codes {
		if !pattern.MatchString(code) {
			t.Fatalf("%s has invalid format: %s", name, code)
		}
		if prev, exists := seen[code]; exists {
			t.Fatalf("duplicate code %s used by %s and %s", code, prev, name)
		}
		seen[code] = name
	}
}
