package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/audit"
)

// AuditMiddleware creates a middleware that logs security-relevant requests
func AuditMiddleware(auditService *audit.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Continue to next handler first
		err := c.Next()

		// Log the request asynchronously (fire and forget)
		// Only log for authenticated requests on security-relevant endpoints
		userID := GetUserID(c)
		if userID != "" {
			// Check if this is a security-relevant endpoint that should be logged
			path := c.Path()
			method := c.Method()

			if shouldAuditRequest(path, method) {
				logSecurityRequest(c, auditService, userID, path, method, err == nil)
			}
		}

		return err
	}
}

// shouldAuditRequest determines if a request should be audited
func shouldAuditRequest(path string, method string) bool {
	// Log all non-GET requests on security endpoints
	if method == "GET" {
		return false
	}

	// List of security-relevant path prefixes to audit
	auditPaths := []string{
		"/api/v1/auth/",
		"/api/v1/providers/",
		"/api/v1/oauth/",
		"/api/v1/github/",
		"/api/v1/mcp/",
		"/api/v1/integrations/",
	}

	for _, prefix := range auditPaths {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

// logSecurityRequest logs a security-relevant request
func logSecurityRequest(c *fiber.Ctx, auditService *audit.Service, userID, path, method string, success bool) {
	// Determine event type based on path and method
	eventType, category := categorizeRequest(path, method)
	if eventType == "" {
		return
	}

	entry := audit.Entry{
		UserID:    &userID,
		EventType: eventType,
		Category:  category,
		Action:    method + " " + path,
		Success:   success,
		Details: map[string]interface{}{
			"method": method,
			"path":   path,
			"status": c.Response().StatusCode(),
		},
	}

	auditService.LogFromRequestAsync(c, entry)
}

// categorizeRequest determines the event type and category for a request
func categorizeRequest(path string, method string) (audit.EventType, audit.EventCategory) {
	// API key operations
	if contains(path, "/providers/") && contains(path, "/key") {
		if method == "POST" {
			return audit.EventProviderKeySet, audit.CategoryProvider
		}
		if method == "DELETE" {
			return audit.EventProviderKeyDeleted, audit.CategoryProvider
		}
	}

	// Auth operations are handled directly in the auth handler
	// This middleware catches any that slip through

	// GitHub operations
	if contains(path, "/github/") || contains(path, "/oauth/github") {
		if method == "DELETE" {
			return audit.EventGitHubDisconnected, audit.CategoryGitHub
		}
	}

	// Settings changes
	if contains(path, "/integrations/") {
		if method == "POST" || method == "DELETE" {
			return audit.EventSettingsChanged, audit.CategorySettings
		}
	}

	return "", ""
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetAuditService retrieves the audit service from the context if set
func GetAuditService(c *fiber.Ctx) *audit.Service {
	if service, ok := c.Locals("auditService").(*audit.Service); ok {
		return service
	}
	return nil
}

// SetAuditService sets the audit service in the context
func SetAuditService(auditService *audit.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("auditService", auditService)
		return c.Next()
	}
}
