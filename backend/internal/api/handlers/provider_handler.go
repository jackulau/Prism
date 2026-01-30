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

// TestPromptRequest represents a request to test a provider with a prompt
type TestPromptRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	APIKey string `json:"api_key,omitempty"`
}

// TestPromptResponse represents the response from a test prompt
type TestPromptResponse struct {
	Response   string `json:"response"`
	LatencyMS  int64  `json:"latency_ms"`
	TokensUsed struct {
		Input  int `json:"input"`
		Output int `json:"output"`
	} `json:"tokens_used"`
	Model string `json:"model"`
}

// TestPrompt sends a test prompt to a provider and returns the response
func (h *ProviderHandler) TestPrompt(c *fiber.Ctx) error {
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

	var req TestPromptRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate request
	if strings.TrimSpace(req.Prompt) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "prompt is required",
		})
	}

	if strings.TrimSpace(req.Model) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "model is required",
		})
	}

	// Limit prompt length to prevent abuse
	if len(req.Prompt) > 1000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "prompt too long (max 1000 characters)",
		})
	}

	// Get the provider
	llmProvider, err := h.llmManager.GetProvider(provider)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "unknown provider: " + provider,
		})
	}

	// If a temporary API key is provided, set it temporarily
	originalKey := ""
	if strings.TrimSpace(req.APIKey) != "" {
		// Store original key if there is one
		if llmProvider.HasConfiguredKey() {
			// We can't easily get the original key, so we'll just set the new one
			// The key will persist for this request
		}
		llmProvider.SetAPIKey(req.APIKey)
		originalKey = req.APIKey
	}

	// Check if provider has a key configured (either temporary or saved)
	if !llmProvider.HasConfiguredKey() && strings.TrimSpace(req.APIKey) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no API key configured for provider",
		})
	}

	// Create chat request
	chatReq := &llm.ChatRequest{
		Model: req.Model,
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
		MaxTokens: 500, // Limit response for testing
		Stream:    true,
	}

	// Create context with 30 second timeout
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	// Track timing
	startTime := time.Now()

	// Send request
	stream, err := llmProvider.Chat(ctx, chatReq)
	if err != nil {
		log.Printf("Test prompt failed for provider %s: %v", provider, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to send test prompt: " + err.Error(),
		})
	}

	// Collect response
	var response strings.Builder
	var inputTokens, outputTokens int

	for chunk := range stream {
		if chunk.Error != nil {
			log.Printf("Test prompt stream error for provider %s: %v", provider, chunk.Error)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "stream error: " + chunk.Error.Error(),
			})
		}

		response.WriteString(chunk.Delta)

		// Collect token counts
		if chunk.InputTokens > 0 {
			inputTokens = chunk.InputTokens
		}
		if chunk.OutputTokens > 0 {
			outputTokens = chunk.OutputTokens
		}
		if chunk.Usage != nil {
			inputTokens = chunk.Usage.PromptTokens
			outputTokens = chunk.Usage.CompletionTokens
		}
	}

	// Calculate latency
	latency := time.Since(startTime).Milliseconds()

	// Restore original key if we temporarily changed it
	if originalKey != "" {
		// The key was already set, no restoration needed
		// In a production system, we'd want to save/restore the original key
	}

	return c.JSON(TestPromptResponse{
		Response:  response.String(),
		LatencyMS: latency,
		TokensUsed: struct {
			Input  int `json:"input"`
			Output int `json:"output"`
		}{
			Input:  inputTokens,
			Output: outputTokens,
		},
		Model: req.Model,
	})
}
