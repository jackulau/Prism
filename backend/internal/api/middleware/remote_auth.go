package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/jacklau/prism/internal/security"
)

// RemoteAuthMiddleware creates middleware for authenticating remote access sessions
func RemoteAuthMiddleware(authService *security.RemoteAuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if remote access is enabled
		if !authService.IsEnabled() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "remote access is not enabled",
			})
		}

		// Get session token from header
		authHeader := c.Get("X-Remote-Session")
		if authHeader == "" {
			// Also check Authorization header with Bearer prefix
			authHeader = c.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					authHeader = parts[1]
				} else {
					authHeader = ""
				}
			}
		}

		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing remote session token",
			})
		}

		// Validate session with IP check
		clientIP := c.IP()
		session, err := authService.ValidateSessionWithIP(authHeader, clientIP)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Store session info in context for downstream handlers
		c.Locals("remoteSession", session)
		c.Locals("remoteIP", session.ClientIP)

		return c.Next()
	}
}

// RemoteAuthRateLimiter creates a rate limiter specifically for remote authenticated requests
func RemoteAuthRateLimiter(maxRequests int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Rate limit by session token if available
			if session, ok := c.Locals("remoteSession").(*security.RemoteSession); ok {
				return "remote:" + session.Token[:16] // Use prefix of token
			}
			return "remote:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate limit exceeded",
				"message": "Too many requests. Please wait before trying again.",
			})
		},
	})
}

// RemoteAuthLoginRateLimiter creates a strict rate limiter for the login endpoint
func RemoteAuthLoginRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10, // 10 attempts per 5 minutes
		Expiration: 5 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "remote-login:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "too many login attempts",
				"message": "Please wait before trying again.",
			})
		},
	})
}

// GetRemoteSession retrieves the remote session from the request context
func GetRemoteSession(c *fiber.Ctx) *security.RemoteSession {
	session, ok := c.Locals("remoteSession").(*security.RemoteSession)
	if !ok {
		return nil
	}
	return session
}

// GetRemoteIP retrieves the remote client IP from the request context
func GetRemoteIP(c *fiber.Ctx) string {
	ip, ok := c.Locals("remoteIP").(string)
	if !ok {
		return ""
	}
	return ip
}

// OptionalRemoteAuthMiddleware creates middleware that allows both authenticated and unauthenticated requests
func OptionalRemoteAuthMiddleware(authService *security.RemoteAuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if remote access is enabled
		if !authService.IsEnabled() {
			return c.Next()
		}

		// Get session token from header
		authHeader := c.Get("X-Remote-Session")
		if authHeader == "" {
			authHeader = c.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					authHeader = parts[1]
				} else {
					authHeader = ""
				}
			}
		}

		if authHeader == "" {
			return c.Next()
		}

		// Try to validate session
		clientIP := c.IP()
		session, err := authService.ValidateSessionWithIP(authHeader, clientIP)
		if err != nil {
			// Don't fail - just continue without session
			return c.Next()
		}

		// Store session info in context
		c.Locals("remoteSession", session)
		c.Locals("remoteIP", session.ClientIP)

		return c.Next()
	}
}
