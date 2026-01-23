package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	mu      sync.Mutex
	tokens  map[string]*bucket
	rps     int
	burst   int
	cleanup time.Duration
}

type bucket struct {
	tokens    float64
	lastCheck time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps, burst int) *RateLimiter {
	rl := &RateLimiter{
		tokens:  make(map[string]*bucket),
		rps:     rps,
		burst:   burst,
		cleanup: 10 * time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Middleware returns a Fiber middleware for rate limiting
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.IP()

		if !rl.allow(key) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests, please try again later",
			})
		}

		return c.Next()
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.tokens[key]
	if !exists {
		b = &bucket{
			tokens:    float64(rl.burst),
			lastCheck: now,
		}
		rl.tokens[key] = b
	}

	// Add tokens based on elapsed time
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += elapsed * float64(rl.rps)
	if b.tokens > float64(rl.burst) {
		b.tokens = float64(rl.burst)
	}
	b.lastCheck = now

	// Check if we have tokens available
	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, b := range rl.tokens {
			if now.Sub(b.lastCheck) > rl.cleanup {
				delete(rl.tokens, key)
			}
		}
		rl.mu.Unlock()
	}
}
