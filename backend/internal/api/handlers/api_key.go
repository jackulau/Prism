package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/security"
)

// Standard API key scopes
const (
	ScopeRead    = "read"    // Read access to conversations, files
	ScopeWrite   = "write"   // Create/modify conversations
	ScopeExecute = "execute" // Code execution
	ScopeAdmin   = "admin"   // Full access
)

// ValidScopes is the list of valid scope values
var ValidScopes = []string{ScopeRead, ScopeWrite, ScopeExecute, ScopeAdmin}

// APIKeyHandler handles user API key management endpoints
type APIKeyHandler struct {
	userAPIKeyRepo  *repository.UserAPIKeyRepository
	providerKeyRepo *repository.ProviderKeyRepository
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(
	userAPIKeyRepo *repository.UserAPIKeyRepository,
	providerKeyRepo *repository.ProviderKeyRepository,
) *APIKeyHandler {
	return &APIKeyHandler{
		userAPIKeyRepo:  userAPIKeyRepo,
		providerKeyRepo: providerKeyRepo,
	}
}

// CreateAPIKeyRequest represents a request to create an API key
type CreateAPIKeyRequest struct {
	Name         string   `json:"name"`
	ExpiresInDays *int    `json:"expires_in_days"`
	Scopes       []string `json:"scopes"`
}

// APIKeyResponse represents API key metadata in responses
type APIKeyResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	Scopes     []string   `json:"scopes"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

// CreateAPIKeyResponse includes the full key (only shown once)
type CreateAPIKeyResponse struct {
	Key string `json:"key"`
	APIKeyResponse
}

// UpdateNameRequest represents a request to update an API key name
type UpdateNameRequest struct {
	Name string `json:"name"`
}

// ListAPIKeys returns all API keys for the authenticated user
func (h *APIKeyHandler) ListAPIKeys(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	keys, err := h.userAPIKeyRepo.GetByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list API keys",
		})
	}

	response := make([]APIKeyResponse, 0, len(keys))
	for _, key := range keys {
		scopes, _ := h.userAPIKeyRepo.GetScopes(key.ID)
		if scopes == nil {
			scopes = []string{}
		}
		response = append(response, APIKeyResponse{
			ID:         key.ID,
			Name:       key.Name,
			Prefix:     key.KeyPrefix,
			Scopes:     scopes,
			CreatedAt:  key.CreatedAt,
			ExpiresAt:  key.ExpiresAt,
			LastUsedAt: key.LastUsedAt,
		})
	}

	return c.JSON(fiber.Map{
		"api_keys": response,
	})
}

// CreateAPIKey creates a new API key for the authenticated user
func (h *APIKeyHandler) CreateAPIKey(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req CreateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	// Validate scopes
	if len(req.Scopes) > 0 {
		if err := ValidateScopes(req.Scopes); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	// Generate the API key
	fullKey, prefix, err := security.GenerateAPIKey("sk")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate API key",
		})
	}

	// Hash the key for storage
	keyHash := security.HashAPIKey(fullKey)

	// Calculate expiration
	var expiresAt *time.Time
	if req.ExpiresInDays != nil && *req.ExpiresInDays > 0 {
		expires := time.Now().AddDate(0, 0, *req.ExpiresInDays)
		expiresAt = &expires
	}

	// Create the key
	apiKey, err := h.userAPIKeyRepo.Create(userID, req.Name, keyHash, prefix, expiresAt, req.Scopes)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create API key",
		})
	}

	// Return the full key (only shown once)
	scopes := req.Scopes
	if scopes == nil {
		scopes = []string{}
	}

	return c.Status(fiber.StatusCreated).JSON(CreateAPIKeyResponse{
		Key: fullKey,
		APIKeyResponse: APIKeyResponse{
			ID:         apiKey.ID,
			Name:       apiKey.Name,
			Prefix:     apiKey.KeyPrefix,
			Scopes:     scopes,
			CreatedAt:  apiKey.CreatedAt,
			ExpiresAt:  apiKey.ExpiresAt,
			LastUsedAt: apiKey.LastUsedAt,
		},
	})
}

// GetAPIKey returns details of a specific API key
func (h *APIKeyHandler) GetAPIKey(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	keyID := c.Params("id")
	if keyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "key id is required",
		})
	}

	apiKey, err := h.userAPIKeyRepo.GetByID(keyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get API key",
		})
	}

	if apiKey == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "API key not found",
		})
	}

	// Verify ownership
	if apiKey.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	scopes, _ := h.userAPIKeyRepo.GetScopes(apiKey.ID)
	if scopes == nil {
		scopes = []string{}
	}

	return c.JSON(APIKeyResponse{
		ID:         apiKey.ID,
		Name:       apiKey.Name,
		Prefix:     apiKey.KeyPrefix,
		Scopes:     scopes,
		CreatedAt:  apiKey.CreatedAt,
		ExpiresAt:  apiKey.ExpiresAt,
		LastUsedAt: apiKey.LastUsedAt,
	})
}

// DeleteAPIKey revokes an API key
func (h *APIKeyHandler) DeleteAPIKey(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	keyID := c.Params("id")
	if keyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "key id is required",
		})
	}

	// Verify ownership
	apiKey, err := h.userAPIKeyRepo.GetByID(keyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get API key",
		})
	}

	if apiKey == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "API key not found",
		})
	}

	if apiKey.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.userAPIKeyRepo.Delete(keyID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete API key",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "API key deleted successfully",
	})
}

// UpdateAPIKeyName updates the name of an API key
func (h *APIKeyHandler) UpdateAPIKeyName(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	keyID := c.Params("id")
	if keyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "key id is required",
		})
	}

	var req UpdateNameRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	// Verify ownership
	apiKey, err := h.userAPIKeyRepo.GetByID(keyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get API key",
		})
	}

	if apiKey == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "API key not found",
		})
	}

	if apiKey.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	if err := h.userAPIKeyRepo.UpdateName(keyID, req.Name); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update API key name",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "API key name updated successfully",
	})
}

// RotateAPIKey rotates an API key (creates new key, invalidates old)
func (h *APIKeyHandler) RotateAPIKey(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	keyID := c.Params("id")
	if keyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "key id is required",
		})
	}

	// Get the existing key
	oldKey, err := h.userAPIKeyRepo.GetByID(keyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get API key",
		})
	}

	if oldKey == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "API key not found",
		})
	}

	if oldKey.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}

	// Get old scopes
	oldScopes, _ := h.userAPIKeyRepo.GetScopes(keyID)

	// Generate new key
	fullKey, prefix, err := security.GenerateAPIKey("sk")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate API key",
		})
	}

	keyHash := security.HashAPIKey(fullKey)

	// Create new key with same name, scopes, and expiration
	newKey, err := h.userAPIKeyRepo.Create(userID, oldKey.Name+" (rotated)", keyHash, prefix, oldKey.ExpiresAt, oldScopes)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create new API key",
		})
	}

	// Delete old key
	if err := h.userAPIKeyRepo.Delete(keyID); err != nil {
		// Log but don't fail - new key is already created
	}

	scopes := oldScopes
	if scopes == nil {
		scopes = []string{}
	}

	return c.Status(fiber.StatusCreated).JSON(CreateAPIKeyResponse{
		Key: fullKey,
		APIKeyResponse: APIKeyResponse{
			ID:         newKey.ID,
			Name:       newKey.Name,
			Prefix:     newKey.KeyPrefix,
			Scopes:     scopes,
			CreatedAt:  newKey.CreatedAt,
			ExpiresAt:  newKey.ExpiresAt,
			LastUsedAt: newKey.LastUsedAt,
		},
	})
}

// ListProviderKeys returns metadata for all provider keys
func (h *APIKeyHandler) ListProviderKeys(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	keys, err := h.providerKeyRepo.GetAllByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list provider keys",
		})
	}

	if keys == nil {
		keys = []repository.ProviderKeyMetadata{}
	}

	return c.JSON(fiber.Map{
		"provider_keys": keys,
	})
}

// GetProviderKeyMetadata returns metadata for a specific provider key
func (h *APIKeyHandler) GetProviderKeyMetadata(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	provider := c.Params("provider")
	if provider == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider is required",
		})
	}

	metadata, err := h.providerKeyRepo.GetKeyMetadata(userID, provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get provider key metadata",
		})
	}

	if metadata == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "provider key not found",
		})
	}

	return c.JSON(metadata)
}

// ValidateScopes checks if all provided scopes are valid
func ValidateScopes(scopes []string) error {
	validSet := make(map[string]bool)
	for _, s := range ValidScopes {
		validSet[s] = true
	}

	for _, scope := range scopes {
		if !validSet[scope] {
			return fiber.NewError(fiber.StatusBadRequest, "invalid scope: "+scope)
		}
	}

	return nil
}

// HasScope checks if a slice of scopes contains a required scope
func HasScope(scopes []string, required string) bool {
	// Admin scope grants all permissions
	for _, s := range scopes {
		if s == ScopeAdmin || s == required {
			return true
		}
	}
	return false
}
