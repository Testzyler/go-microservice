package httpx

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/Testzyler/go-microservice/pkg/auth"
	"github.com/Testzyler/go-microservice/pkg/token"
)

type Config struct {
	TokenService token.Service
	Optional     bool
}

// Middleware validates Bearer JWTs and stores claims in context.
func Middleware(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			if cfg.Optional {
				return c.Next()
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization header format"})
		}
		tokenString := parts[1]

		claims, err := cfg.TokenService.ParseAccessToken(context.Background(), tokenString)
		if err != nil {
			if cfg.Optional {
				return c.Next()
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}
		c.Locals("user_id", claims.Subject.String())
		c.Locals("claims", &claims)
		return c.Next()
	}
}

func Claims(c *fiber.Ctx) *auth.Claims {
	if claims, ok := c.Locals("claims").(*auth.Claims); ok {
		return claims
	}
	return nil
}

func UserID(c *fiber.Ctx) string {
	if id, ok := c.Locals("user_id").(string); ok {
		return id
	}
	return ""
}
