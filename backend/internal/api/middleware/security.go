package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders() fiber.Handler {
	isProduction := os.Getenv("ENVIRONMENT") == "production"

	return func(c *fiber.Ctx) error {
		// Prevent MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking - use SAMEORIGIN to allow preview iframes
		c.Set("X-Frame-Options", "SAMEORIGIN")

		// XSS protection
		c.Set("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy
		// Note: Monaco editor requires 'unsafe-eval' for web workers
		// 'unsafe-inline' is needed for inline styles
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: blob: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self' ws: wss:; " +
			"frame-src 'self';"
		c.Set("Content-Security-Policy", csp)

		// Strict Transport Security - enable in production with HTTPS
		if isProduction || c.Protocol() == "https" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		return c.Next()
	}
}
