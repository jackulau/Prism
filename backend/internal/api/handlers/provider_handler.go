package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/security"
)

// ProviderHandler handles provider key management endpoints
type ProviderHandler struct {
	providerKeyRepo   *repository.ProviderKeyRepository
	encryptionService *security.EncryptionService
	llmManager        *llm.Manager
}

// NewProviderHandler creates a new provider handler
func NewProviderHandler(
	providerKeyRepo *repository.ProviderKeyRepository,
	encryptionService *security.EncryptionService,
	llmManager *llm.Manager,
) *ProviderHandler {
	return &ProviderHandler{
		providerKeyRepo:   providerKeyRepo,
		encryptionService: encryptionService,
		llmManager:        llmManager,
	}
}

// SetKeyRequest represents a request to set an API key
type SetKeyRequest struct {
	APIKey string `json:"api_key"`
}

// SetKey stores an encrypted API key for a provider
func (h *ProviderHandler) SetKey(c *fiber.Ctx) error {
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

	// Validate provider exists
	if _, err := h.llmManager.GetProvider(provider); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "unknown provider: " + provider,
		})
	}

	var req SetKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if strings.TrimSpace(req.APIKey) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "api_key is required",
		})
	}

	// Encrypt the API key
	encryptedKey, nonce, err := h.encryptionService.Encrypt([]byte(req.APIKey))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to encrypt API key",
		})
	}

	// Store the encrypted key
	if err := h.providerKeyRepo.SetKey(userID, provider, encryptedKey, nonce); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save API key",
		})
	}

	// Also set the key on the provider instance for immediate use
	h.llmManager.SetAPIKey(provider, req.APIKey)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "API key saved successfully",
	})
}

// DeleteKey removes an API key for a provider
func (h *ProviderHandler) DeleteKey(c *fiber.Ctx) error {
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

	if err := h.providerKeyRepo.DeleteKey(userID, provider); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete API key",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "API key deleted successfully",
	})
}

// ValidateKeyRequest represents a request to validate an API key
type ValidateKeyRequest struct {
	APIKey string `json:"api_key"`
}

// ValidateKey validates an API key with the provider
func (h *ProviderHandler) ValidateKey(c *fiber.Ctx) error {
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

	var req ValidateKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if strings.TrimSpace(req.APIKey) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "api_key is required",
		})
	}

	// Validate the API key by making a test request to the provider
	valid, err := h.validateProviderKey(provider, req.APIKey)
	if err != nil {
		log.Printf("Provider key validation failed for %s: %v", provider, err)
		return c.JSON(fiber.Map{
			"valid":   false,
			"message": "API key validation failed",
		})
	}

	return c.JSON(fiber.Map{
		"valid": valid,
	})
}

// GetKeyStatus returns whether a user has a key configured for a provider
func (h *ProviderHandler) GetKeyStatus(c *fiber.Ctx) error {
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

	hasKey, err := h.providerKeyRepo.HasKey(userID, provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check API key status",
		})
	}

	return c.JSON(fiber.Map{
		"has_key":  hasKey,
		"provider": provider,
	})
}

// ListKeys returns a list of providers the user has keys configured for
func (h *ProviderHandler) ListKeys(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	keys, err := h.providerKeyRepo.ListKeys(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list API keys",
		})
	}

	providers := make([]string, len(keys))
	for i, key := range keys {
		providers[i] = key.Provider
	}

	return c.JSON(fiber.Map{
		"providers": providers,
	})
}

// validateProviderKey validates an API key by making a simple API call
func (h *ProviderHandler) validateProviderKey(provider, apiKey string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch provider {
	case "openai":
		return h.validateOpenAIKey(ctx, apiKey)
	case "anthropic":
		return h.validateAnthropicKey(ctx, apiKey)
	case "ollama":
		// Ollama doesn't require an API key
		return true, nil
	default:
		// For unknown providers, just check if the key is non-empty
		return len(apiKey) > 0, nil
	}
}

// validateOpenAIKey validates an OpenAI API key
func (h *ProviderHandler) validateOpenAIKey(ctx context.Context, apiKey string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// validateAnthropicKey validates an Anthropic API key
func (h *ProviderHandler) validateAnthropicKey(ctx context.Context, apiKey string) (bool, error) {
	// Anthropic doesn't have a simple list models endpoint
	// We'll check the key format and make a minimal request
	if !strings.HasPrefix(apiKey, "sk-ant-") {
		return false, nil
	}

	// Make a minimal request to check if the key is valid
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.anthropic.com/v1/messages", nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// 401 means invalid key, 405 means valid key but wrong method (which is fine)
	return resp.StatusCode != http.StatusUnauthorized, nil
}
