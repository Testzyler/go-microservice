package proxy

import (
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	authv1 "github.com/Testzyler/go-microservice/gen/proto/auth/v1"
	authv1connect "github.com/Testzyler/go-microservice/gen/proto/auth/v1/authv1connect"
	"github.com/Testzyler/go-microservice/global/pkg/auth"
	"github.com/Testzyler/go-microservice/global/pkg/token"
	"github.com/Testzyler/go-microservice/services/internal-gateway/internal/config"
)

type authHandlers struct {
	auth   authv1connect.AuthServiceClient
	tokens *token.JWTService
	logger *zap.Logger
	cfg    *config.Config
}

func newAuthHandlers(authClient authv1connect.AuthServiceClient, tokens *token.JWTService, logger *zap.Logger, cfg *config.Config) *authHandlers {
	return &authHandlers{auth: authClient, tokens: tokens, logger: logger, cfg: cfg}
}

func (h *authHandlers) issueToken(c *fiber.Ctx) error {
	var req struct {
		UserID      string   `json:"user_id"`
		Roles       []string `json:"roles"`
		Permissions []string `json:"permissions"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id is required"})
	}
	subject, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id must be a valid UUID"})
	}

	claims := auth.Claims{
		Subject:     subject,
		Roles:       append([]string(nil), req.Roles...),
		Permissions: append([]string(nil), req.Permissions...),
	}

	token, err := h.tokens.IssueAccessToken(c.UserContext(), claims)
	if err != nil {
		h.logger.Error("failed to issue token", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to issue token"})
	}

	return c.JSON(fiber.Map{
		"access_token": token,
		"expires_in":   int64(h.cfg.AccessTokenTTL.Seconds()),
		"issued_at":    time.Now().UTC(),
	})
}

func (h *authHandlers) getCurrentUser(c *fiber.Ctx) error {
	claims := claimsFromLocals(c)
	if claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing claims"})
	}
	return c.JSON(fiber.Map{
		"user_id":     claims.Subject.String(),
		"roles":       claims.Roles,
		"permissions": claims.Permissions,
	})
}

func (h *authHandlers) register(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	resp, err := h.auth.Register(c.UserContext(), connect.NewRequest(&authv1.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	}))
	if err != nil {
		return writeConnectError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp.Msg)
}

func (h *authHandlers) login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	resp, err := h.auth.Login(c.UserContext(), connect.NewRequest(&authv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}))
	if err != nil {
		return writeConnectError(c, err)
	}
	return c.JSON(resp.Msg)
}

func (h *authHandlers) authValidate(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization header format"})
	}
	tokenString := parts[1]

	resp, err := h.auth.Validate(c.UserContext(), connect.NewRequest(&authv1.ValidateRequest{Token: tokenString}))
	if err != nil {
		return writeConnectError(c, err)
	}
	if !resp.Msg.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
	}
	claims := &auth.Claims{
		Subject:     uuid.MustParse(resp.Msg.UserId),
		Roles:       append([]string(nil), resp.Msg.Roles...),
		Permissions: append([]string(nil), resp.Msg.Permissions...),
	}
	c.Locals("claims", claims)
	c.Locals("user_id", resp.Msg.UserId)
	return c.Next()
}

func claimsFromLocals(c *fiber.Ctx) *auth.Claims {
	if claims, ok := c.Locals("claims").(*auth.Claims); ok {
		return claims
	}
	return nil
}
