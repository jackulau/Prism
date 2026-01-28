package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/llm/openaicompat"
	"github.com/jacklau/prism/internal/security"
)

// CustomProviderHandler handles custom OpenAI-compatible provider endpoints
type CustomProviderHandler struct {
	customProviderRepo *repository.CustomProviderRepository
	encryptionService  *security.EncryptionService
	llmManager         *llm.Manager
}

// NewCustomProviderHandler creates a new custom provider handler
func NewCustomProviderHandler(
	customProviderRepo *repository.CustomProviderRepository,
	encryptionService *security.EncryptionService,
	llmManager *llm.Manager,
) *CustomProviderHandler {
	return &CustomProviderHandler{
		customProviderRepo: customProviderRepo,
		encryptionService:  encryptionService,
		llmManager:         llmManager,
	}
}

// CreateCustomProviderRequest represents a request to create a custom provider
type CreateCustomProviderRequest struct {
	Name           string   `json:"name"`
	BaseURL        string   `json:"base_url"`
	APIKey         string   `json:"api_key,omitempty"`
	Models         []string `json:"models,omitempty"`
	SupportsTools  bool     `json:"supports_tools"`
	SupportsVision bool     `json:"supports_vision"`
}

// UpdateCustomProviderRequest represents a request to update a custom provider
type UpdateCustomProviderRequest struct {
	Name           *string  `json:"name,omitempty"`
	BaseURL        *string  `json:"base_url,omitempty"`
	APIKey         *string  `json:"api_key,omitempty"`
	Models         []string `json:"models,omitempty"`
	SupportsTools  *bool    `json:"supports_tools,omitempty"`
	SupportsVision *bool    `json:"supports_vision,omitempty"`
}

// CustomProviderResponse represents a custom provider in API responses
type CustomProviderResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	BaseURL        string    `json:"base_url"`
	HasAPIKey      bool      `json:"has_api_key"`
	Models         []string  `json:"models"`
	SupportsTools  bool      `json:"supports_tools"`
	SupportsVision bool      `json:"supports_vision"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TestEndpointRequest represents a request to test an endpoint
type TestEndpointRequest struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key,omitempty"`
}

// Create creates a new custom provider
func (h *CustomProviderHandler) Create(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req CreateCustomProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate required fields
	if strings.TrimSpace(req.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	if strings.TrimSpace(req.BaseURL) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "base_url is required",
		})
	}

	// Validate name doesn't conflict with built-in providers
	reservedNames := []string{"openai", "anthropic", "ollama", "google", "gemini"}
	nameLower := strings.ToLower(req.Name)
	for _, reserved := range reservedNames {
		if nameLower == reserved {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "name '" + req.Name + "' is reserved for built-in providers",
			})
		}
	}

	// Check if provider with same name already exists
	exists, err := h.customProviderRepo.Exists(userID, req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check provider existence",
		})
	}
	if exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "a provider with this name already exists",
		})
	}

	// Convert models to JSON
	var modelsJSON string
	if len(req.Models) > 0 {
		modelsBytes, err := json.Marshal(req.Models)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to encode models",
			})
		}
		modelsJSON = string(modelsBytes)
	} else {
		modelsJSON = "[]"
	}

	// Create the provider
	provider, err := h.customProviderRepo.Create(
		userID,
		req.Name,
		req.BaseURL,
		req.APIKey,
		modelsJSON,
		req.SupportsTools,
		req.SupportsVision,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create custom provider",
		})
	}

	// Register the provider with the LLM manager
	h.registerProvider(provider, req.APIKey)

	// Parse models for response
	var models []string
	if err := json.Unmarshal([]byte(modelsJSON), &models); err != nil {
		models = []string{}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"provider": CustomProviderResponse{
			ID:             provider.ID,
			Name:           provider.Name,
			BaseURL:        provider.BaseURL,
			HasAPIKey:      provider.EncryptedKey != nil && len(provider.EncryptedKey) > 0,
			Models:         models,
			SupportsTools:  provider.SupportsTools,
			SupportsVision: provider.SupportsVision,
			IsActive:       provider.IsActive,
			CreatedAt:      provider.CreatedAt,
			UpdatedAt:      provider.UpdatedAt,
		},
	})
}

// List returns all custom providers for the user
func (h *CustomProviderHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	providers, err := h.customProviderRepo.List(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list custom providers",
		})
	}

	responses := make([]CustomProviderResponse, len(providers))
	for i, p := range providers {
		var models []string
		if p.Models != "" {
			if err := json.Unmarshal([]byte(p.Models), &models); err != nil {
				models = []string{}
			}
		}

		responses[i] = CustomProviderResponse{
			ID:             p.ID,
			Name:           p.Name,
			BaseURL:        p.BaseURL,
			HasAPIKey:      p.HasAPIKey,
			Models:         models,
			SupportsTools:  p.SupportsTools,
			SupportsVision: p.SupportsVision,
			IsActive:       p.IsActive,
			CreatedAt:      p.CreatedAt,
			UpdatedAt:      p.UpdatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"providers": responses,
	})
}

// Get returns a specific custom provider
func (h *CustomProviderHandler) Get(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	providerID := c.Params("id")
	if providerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider id is required",
		})
	}

	provider, err := h.customProviderRepo.GetByID(providerID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get custom provider",
		})
	}
	if provider == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "custom provider not found",
		})
	}

	var models []string
	if provider.Models != "" {
		if err := json.Unmarshal([]byte(provider.Models), &models); err != nil {
			models = []string{}
		}
	}

	return c.JSON(fiber.Map{
		"provider": CustomProviderResponse{
			ID:             provider.ID,
			Name:           provider.Name,
			BaseURL:        provider.BaseURL,
			HasAPIKey:      provider.EncryptedKey != nil && len(provider.EncryptedKey) > 0,
			Models:         models,
			SupportsTools:  provider.SupportsTools,
			SupportsVision: provider.SupportsVision,
			IsActive:       provider.IsActive,
			CreatedAt:      provider.CreatedAt,
			UpdatedAt:      provider.UpdatedAt,
		},
	})
}

// Update updates a custom provider
func (h *CustomProviderHandler) Update(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	providerID := c.Params("id")
	if providerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider id is required",
		})
	}

	var req UpdateCustomProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Get existing provider to check ownership
	existing, err := h.customProviderRepo.GetByID(providerID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get custom provider",
		})
	}
	if existing == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "custom provider not found",
		})
	}

	// Validate name if provided
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "name cannot be empty",
			})
		}

		// Check for reserved names
		reservedNames := []string{"openai", "anthropic", "ollama", "google", "gemini"}
		nameLower := strings.ToLower(*req.Name)
		for _, reserved := range reservedNames {
			if nameLower == reserved {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "name '" + *req.Name + "' is reserved for built-in providers",
				})
			}
		}

		// Check if name is already taken by another provider
		if *req.Name != existing.Name {
			exists, err := h.customProviderRepo.Exists(userID, *req.Name)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to check provider existence",
				})
			}
			if exists {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "a provider with this name already exists",
				})
			}
		}
	}

	// Convert models to JSON if provided
	var modelsJSON *string
	if req.Models != nil {
		modelsBytes, err := json.Marshal(req.Models)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to encode models",
			})
		}
		modelsStr := string(modelsBytes)
		modelsJSON = &modelsStr
	}

	// Update the provider
	err = h.customProviderRepo.Update(
		providerID,
		userID,
		req.Name,
		req.BaseURL,
		req.APIKey,
		modelsJSON,
		req.SupportsTools,
		req.SupportsVision,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update custom provider",
		})
	}

	// Get updated provider
	updated, err := h.customProviderRepo.GetByID(providerID, userID)
	if err != nil || updated == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get updated provider",
		})
	}

	// Re-register the provider if key was updated
	if req.APIKey != nil {
		h.registerProvider(updated, *req.APIKey)
	}

	var models []string
	if updated.Models != "" {
		if err := json.Unmarshal([]byte(updated.Models), &models); err != nil {
			models = []string{}
		}
	}

	return c.JSON(fiber.Map{
		"provider": CustomProviderResponse{
			ID:             updated.ID,
			Name:           updated.Name,
			BaseURL:        updated.BaseURL,
			HasAPIKey:      updated.EncryptedKey != nil && len(updated.EncryptedKey) > 0,
			Models:         models,
			SupportsTools:  updated.SupportsTools,
			SupportsVision: updated.SupportsVision,
			IsActive:       updated.IsActive,
			CreatedAt:      updated.CreatedAt,
			UpdatedAt:      updated.UpdatedAt,
		},
	})
}

// Delete removes a custom provider
func (h *CustomProviderHandler) Delete(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	providerID := c.Params("id")
	if providerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider id is required",
		})
	}

	// Get provider to unregister from LLM manager
	provider, err := h.customProviderRepo.GetByID(providerID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get custom provider",
		})
	}
	if provider == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "custom provider not found",
		})
	}

	// Delete the provider
	err = h.customProviderRepo.Delete(providerID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete custom provider",
		})
	}

	// Note: We don't unregister from LLM manager as it would affect active conversations
	// The provider will be cleaned up on next server restart

	return c.JSON(fiber.Map{
		"success": true,
		"message": "custom provider deleted successfully",
	})
}

// TestEndpoint tests connectivity to an OpenAI-compatible endpoint
func (h *CustomProviderHandler) TestEndpoint(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req TestEndpointRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if strings.TrimSpace(req.BaseURL) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "base_url is required",
		})
	}

	// Create a model fetcher to test the endpoint
	fetcher := openaicompat.NewModelFetcher()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := fetcher.TestEndpoint(ctx, req.BaseURL, req.APIKey)
	if err != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success":     false,
			"accessible":  false,
			"message":     err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":          result.Accessible,
		"accessible":       result.Accessible,
		"models_available": result.ModelsAvailable,
		"auth_required":    result.AuthRequired,
		"model_count":      result.ModelCount,
		"message":          result.Message,
	})
}

// FetchModels fetches available models from a custom provider
func (h *CustomProviderHandler) FetchModels(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	providerID := c.Params("id")
	if providerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider id is required",
		})
	}

	provider, err := h.customProviderRepo.GetByID(providerID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get custom provider",
		})
	}
	if provider == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "custom provider not found",
		})
	}

	// Decrypt API key if present
	apiKey, err := h.customProviderRepo.DecryptAPIKey(provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to decrypt API key",
		})
	}

	// Fetch models
	fetcher := openaicompat.NewModelFetcher()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	models, err := fetcher.FetchModels(ctx, provider.BaseURL, apiKey)
	if err != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Convert to model IDs for storage
	modelIDs := make([]string, len(models))
	for i, m := range models {
		modelIDs[i] = m.ID
	}

	// Update the provider with fetched models
	modelsBytes, _ := json.Marshal(modelIDs)
	modelsStr := string(modelsBytes)
	err = h.customProviderRepo.Update(providerID, userID, nil, nil, nil, &modelsStr, nil, nil)
	if err != nil {
		// Log but don't fail - we still return the models
	}

	return c.JSON(fiber.Map{
		"success": true,
		"models":  models,
	})
}

// registerProvider registers a custom provider with the LLM manager
func (h *CustomProviderHandler) registerProvider(provider *repository.CustomProvider, apiKey string) {
	// Parse models from JSON
	var modelIDs []string
	if provider.Models != "" {
		json.Unmarshal([]byte(provider.Models), &modelIDs)
	}

	// Convert to llm.Model
	models := make([]llm.Model, len(modelIDs))
	for i, id := range modelIDs {
		models[i] = llm.Model{
			ID:             id,
			Name:           id,
			ContextWindow:  4096, // Default, will be updated when fetched
			SupportsTools:  provider.SupportsTools,
			SupportsVision: provider.SupportsVision,
		}
	}

	// Create the openaicompat client
	cfg := openaicompat.Config{
		ID:             provider.ID,
		Name:           provider.Name,
		BaseURL:        provider.BaseURL,
		APIKey:         apiKey,
		SupportsTools:  provider.SupportsTools,
		SupportsVision: provider.SupportsVision,
		Models:         models,
	}

	client := openaicompat.NewClient(cfg)
	h.llmManager.RegisterProvider(client)
}
