package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/security"
)

// WorkOSSessionMiddleware validates wos-session cookie and sets org context
func WorkOSSessionMiddleware(workosService *security.WorkOSService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check for wos-session cookie
		cookie := c.Cookies("wos-session")
		if cookie == "" {
			return c.Next() // No session, continue
		}

		// Decrypt and validate session
		session, err := workosService.DecryptSession(cookie)
		if err != nil {
			// Invalid session, clear cookie and continue
			c.Cookie(&fiber.Cookie{
				Name:     "wos-session",
				Value:    "",
				Expires:  time.Now().Add(-time.Hour),
				HTTPOnly: true,
				Secure:   true,
				SameSite: "Lax",
			})
			return c.Next()
		}

		// Check expiration
		if session.ExpiresAt.Before(time.Now()) {
			// Expired session, clear cookie and continue
			c.Cookie(&fiber.Cookie{
				Name:     "wos-session",
				Value:    "",
				Expires:  time.Now().Add(-time.Hour),
				HTTPOnly: true,
				Secure:   true,
				SameSite: "Lax",
			})
			return c.Next()
		}

		// Set organization context
		c.Locals("organizationID", session.OrganizationID)
		c.Locals("ssoSessionID", session.ID)
		c.Locals("ssoUserID", session.UserID)
		c.Locals("connectionID", session.ConnectionID)

		return c.Next()
	}
}

// GetOrganizationID returns the organization ID from context
func GetOrganizationID(c *fiber.Ctx) string {
	if orgID, ok := c.Locals("organizationID").(string); ok {
		return orgID
	}
	return ""
}

// HasOrganizationContext checks if request has org context
func HasOrganizationContext(c *fiber.Ctx) bool {
	return GetOrganizationID(c) != ""
}

// GetSSOSessionID returns the SSO session ID from context
func GetSSOSessionID(c *fiber.Ctx) string {
	if sessionID, ok := c.Locals("ssoSessionID").(string); ok {
		return sessionID
	}
	return ""
}

// GetSSOUserID returns the SSO user ID from context
func GetSSOUserID(c *fiber.Ctx) string {
	if userID, ok := c.Locals("ssoUserID").(string); ok {
		return userID
	}
	return ""
}

// GetConnectionID returns the connection ID from context
func GetConnectionID(c *fiber.Ctx) string {
	if connID, ok := c.Locals("connectionID").(string); ok {
		return connID
	}
	return ""
}

// RequireOrganization ensures request has organization context
func RequireOrganization() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !HasOrganizationContext(c) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Organization context required",
			})
		}
		return c.Next()
	}
}
