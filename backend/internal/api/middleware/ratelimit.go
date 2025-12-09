package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimiter creates a rate limiting middleware with the given configuration
func RateLimiter(requestsPerMinute, burst int) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        requestsPerMinute,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use user ID if authenticated, otherwise use IP
			if userID := c.Locals("userID"); userID != nil {
				return userID.(string)
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate limit exceeded",
				"message": "Too many requests. Please wait before trying again.",
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
	})
}

// StrictRateLimiter creates a stricter rate limiter for sensitive endpoints
func StrictRateLimiter(maxRequests int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate limit exceeded",
				"message": "Too many attempts. Please wait before trying again.",
			})
		},
	})
}
