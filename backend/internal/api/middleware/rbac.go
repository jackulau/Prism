package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/security"
)

const (
	// UserRoleKey is the context key for storing user role
	UserRoleKey = "userRole"
)

// RequireRole creates a middleware that checks if the user has one of the required roles
func RequireRole(roles ...security.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := GetUserRole(c)
		if userRole == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied: no role assigned",
			})
		}

		for _, role := range roles {
			if userRole == string(role) {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient permissions",
		})
	}
}

// RequirePermission creates a middleware that checks if the user has a specific permission
func RequirePermission(perm security.Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := security.Role(GetUserRole(c))
		if userRole == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied: no role assigned",
			})
		}

		if !security.HasPermission(userRole, perm) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireAdmin is a convenience middleware that requires the admin role
func RequireAdmin() fiber.Handler {
	return RequireRole(security.RoleAdmin)
}

// GetUserRole retrieves the user role from the fiber context
func GetUserRole(c *fiber.Ctx) string {
	role, ok := c.Locals(UserRoleKey).(string)
	if !ok {
		return ""
	}
	return role
}

// SetUserRole stores the user role in the fiber context
func SetUserRole(c *fiber.Ctx, role string) {
	c.Locals(UserRoleKey, role)
}

// IsAdmin checks if the current user has admin role
func IsAdmin(c *fiber.Ctx) bool {
	return GetUserRole(c) == string(security.RoleAdmin)
}

// HasPermission checks if the current user has a specific permission
func HasPermission(c *fiber.Ctx, perm security.Permission) bool {
	role := security.Role(GetUserRole(c))
	return security.HasPermission(role, perm)
}
