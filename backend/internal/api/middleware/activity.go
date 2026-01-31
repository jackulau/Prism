package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/services/session"
)

// ActivityMiddleware creates a middleware that tracks session activity
func ActivityMiddleware(sessionService *session.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := GetSessionID(c)
		if sessionID != "" {
			// Update activity asynchronously to not block the request
			go func(sid string) {
				_ = sessionService.RecordActivity(sid)
			}(sessionID)
		}
		return c.Next()
	}
}

// GetSessionID gets the session ID from the context
func GetSessionID(c *fiber.Ctx) string {
	sessionID, ok := c.Locals("sessionID").(string)
	if !ok {
		return ""
	}
	return sessionID
}
