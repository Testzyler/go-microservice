package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(cfg CORSConfig) fiber.Handler {
	allowOrigins := "*"
	allowCredentials := false

	// Only enable credentials when specific origins are configured (not wildcard)
	if len(cfg.AllowOrigins) > 0 {
		allowOrigins = join(cfg.AllowOrigins, ",")
		// Don't allow credentials with wildcard origin (Fiber security restriction)
		if allowOrigins != "*" {
			allowCredentials = true
		}
	}

	allowMethods := "GET,POST,PUT,DELETE,OPTIONS"
	if len(cfg.AllowMethods) > 0 {
		allowMethods = join(cfg.AllowMethods, ",")
	}

	allowHeaders := "Origin,Content-Type,Accept,Authorization"
	if len(cfg.AllowHeaders) > 0 {
		allowHeaders = join(cfg.AllowHeaders, ",")
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		AllowCredentials: allowCredentials,
		ExposeHeaders:    "Content-Length,Content-Type",
		MaxAge:           86400,
	})
}

func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
