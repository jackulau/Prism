package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/security"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(jwtService *security.JWTService) fiber.Handler {
	return AuthMiddlewareWithRole(jwtService, nil)
}

// AuthMiddlewareWithRole creates a middleware for JWT authentication that also fetches user role
func AuthMiddlewareWithRole(jwtService *security.JWTService, userRepo *repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		// Check for Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		token := parts[1]

		// Validate token
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		// Set user info in context
		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)

		// Fetch and set user role if repository is provided
		if userRepo != nil {
			role, err := userRepo.GetUserRole(claims.UserID)
			if err == nil {
				SetUserRole(c, role)
			} else {
				// Default to "user" role if role fetch fails
				SetUserRole(c, "user")
			}
		}

		return c.Next()
	}
}

// OptionalAuthMiddleware creates a middleware that allows both authenticated and unauthenticated requests
func OptionalAuthMiddleware(jwtService *security.JWTService) fiber.Handler {
	return OptionalAuthMiddlewareWithRole(jwtService, nil)
}

// OptionalAuthMiddlewareWithRole creates optional auth middleware that also fetches user role
func OptionalAuthMiddlewareWithRole(jwtService *security.JWTService, userRepo *repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Next()
		}

		token := parts[1]
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			return c.Next()
		}

		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)

		// Fetch and set user role if repository is provided
		if userRepo != nil {
			role, err := userRepo.GetUserRole(claims.UserID)
			if err == nil {
				SetUserRole(c, role)
			} else {
				SetUserRole(c, "user")
			}
		}

		return c.Next()
	}
}

// GetUserID gets the user ID from the context
func GetUserID(c *fiber.Ctx) string {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return ""
	}
	return userID
}

// GetEmail gets the email from the context
func GetEmail(c *fiber.Ctx) string {
	email, ok := c.Locals("email").(string)
	if !ok {
		return ""
	}
	return email
}

// GetAPIKeyScopes gets the API key scopes from the context (set by APIKeyAuthMiddleware)
func GetAPIKeyScopes(c *fiber.Ctx) []string {
	scopes, ok := c.Locals("apiKeyScopes").([]string)
	if !ok {
		return nil
	}
	return scopes
}

// APIKeyAuthMiddleware creates a middleware for API key authentication
// It validates API keys passed via the X-API-Key header
func APIKeyAuthMiddleware(apiKeyRepo *repository.UserAPIKeyRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check X-API-Key header
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing X-API-Key header",
			})
		}

		// Hash the provided key
		keyHash := security.HashAPIKey(apiKey)

		// Look up the key
		key, err := apiKeyRepo.GetByKeyHash(keyHash)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to validate API key",
			})
		}

		if key == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid API key",
			})
		}

		// Check expiration
		if key.IsExpired() {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "API key has expired",
			})
		}

		// Update last used timestamp (async to not block request)
		go func() {
			_ = apiKeyRepo.UpdateLastUsed(key.ID)
		}()

		// Get scopes
		scopes, _ := apiKeyRepo.GetScopes(key.ID)

		// Set user context
		c.Locals("userID", key.UserID)
		c.Locals("apiKeyID", key.ID)
		c.Locals("apiKeyScopes", scopes)

		return c.Next()
	}
}

// APIKeyOrJWTAuthMiddleware creates a middleware that accepts either API key or JWT auth
func APIKeyOrJWTAuthMiddleware(jwtService *security.JWTService, apiKeyRepo *repository.UserAPIKeyRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First try X-API-Key header
		apiKey := c.Get("X-API-Key")
		if apiKey != "" {
			keyHash := security.HashAPIKey(apiKey)
			key, err := apiKeyRepo.GetByKeyHash(keyHash)
			if err == nil && key != nil && !key.IsExpired() {
				// Update last used
				go func() {
					_ = apiKeyRepo.UpdateLastUsed(key.ID)
				}()

				scopes, _ := apiKeyRepo.GetScopes(key.ID)
				c.Locals("userID", key.UserID)
				c.Locals("apiKeyID", key.ID)
				c.Locals("apiKeyScopes", scopes)
				return c.Next()
			}
		}

		// Fall back to JWT auth
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		token := parts[1]
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}

// RequireScope creates a middleware that checks if the request has a required scope
// This should be used after APIKeyAuthMiddleware or APIKeyOrJWTAuthMiddleware
func RequireScope(requiredScope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If authenticated via JWT (no API key), allow all scopes
		apiKeyID := c.Locals("apiKeyID")
		if apiKeyID == nil {
			return c.Next()
		}

		// Check scopes for API key auth
		scopes := GetAPIKeyScopes(c)
		if !hasScope(scopes, requiredScope) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "insufficient scope: " + requiredScope + " required",
			})
		}

		return c.Next()
	}
}

// hasScope checks if a slice of scopes contains a required scope
func hasScope(scopes []string, required string) bool {
	for _, s := range scopes {
		// Admin scope grants all permissions
		if s == "admin" || s == required {
			return true
		}
	}
	return false
}
