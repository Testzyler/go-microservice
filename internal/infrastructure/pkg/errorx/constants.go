package errorx

// Error code format:
// <CLASS>-<DOMAIN>-<HTTP_STATUS>-<SEQUENCE>
//
// CLASS
// - SYS: technical/system error
// - BIZ: business/domain error
//
// DOMAIN
// - GEN: generic
// - HTTP: transport boundary
// - IDP: idempotency domain
// - SLR: seller domain
// - PRD: product domain
//
// Example: BIZ-SLR-404-001
const (
	CodeInternal = "SYS-GEN-500-001"

	MessageInternalServerError = "internal server error"
)

const (
	CodeValidationError  = "BIZ-GEN-400-001"
	MessageValidationErr = "validation failed"

	CodeHTTPBadRequest        = "SYS-HTTP-400-001"
	MessageInvalidRequestBody = "invalid request body"
)

const (
	CodeIdempotencyInProgress    = "BIZ-IDP-409-001"
	MessageIdempotencyInProgress = "request with same idempotency key is still in progress"
	CodeIdempotencyKeyReused     = "BIZ-IDP-409-002"
	MessageIdempotencyKeyReused  = "idempotency key is already used with different request payload"
)

const (
	CodeSellerNotFound         = "BIZ-SLR-404-001"
	MessageSellerNotFound      = "seller not found"
	CodeSellerInvalidID        = "BIZ-SLR-400-001"
	MessageSellerInvalidID     = "invalid seller id"
	CodeSellerIDRequired       = "BIZ-SLR-400-002"
	MessageSellerIDRequired    = "seller id is required"
	CodeSellerQueryRequired    = "BIZ-SLR-400-003"
	MessageSellerQueryRequired = "seller query is required"
)

const (
	CodeProductNotFound           = "BIZ-PRD-404-001"
	MessageProductNotFound        = "product not found"
	CodeProductInvalidID          = "BIZ-PRD-400-001"
	MessageProductInvalidID       = "invalid product id"
	CodeProductIDRequired         = "BIZ-PRD-400-002"
	MessageProductIDRequired      = "product id is required"
	CodeProductQueryRequired      = "BIZ-PRD-400-003"
	MessageProductQueryRequired   = "product query is required"
	CodeProductInvalidSellerID    = "BIZ-PRD-400-004"
	MessageProductInvalidSellerID = "sellerId must be a valid UUID"
)
