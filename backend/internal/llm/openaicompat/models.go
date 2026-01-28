package openaicompat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jacklau/prism/internal/llm"
)

// ModelFetcher handles fetching and managing models from OpenAI-compatible endpoints
type ModelFetcher struct {
	client  *http.Client
	timeout time.Duration
}

// NewModelFetcher creates a new ModelFetcher
func NewModelFetcher() *ModelFetcher {
	return &ModelFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30 * time.Second,
	}
}

// FetchModels fetches available models from an OpenAI-compatible endpoint
func (f *ModelFetcher) FetchModels(ctx context.Context, baseURL, apiKey string) ([]llm.Model, error) {
	baseURL = strings.TrimSuffix(baseURL, "/")
	url := baseURL + "/v1/models"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("models endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var modelsResp ModelsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to parse models response: %w", err)
	}

	models := make([]llm.Model, len(modelsResp.Data))
	for i, m := range modelsResp.Data {
		models[i] = convertAPIModel(m)
	}

	return models, nil
}

// TestEndpoint tests if an endpoint is accessible and optionally validates the API key
func (f *ModelFetcher) TestEndpoint(ctx context.Context, baseURL, apiKey string) (*EndpointTestResult, error) {
	baseURL = strings.TrimSuffix(baseURL, "/")
	result := &EndpointTestResult{
		Accessible:    false,
		ModelsAvailable: false,
	}

	// Try the models endpoint first
	modelsURL := baseURL + "/v1/models"
	req, err := http.NewRequestWithContext(ctx, "GET", modelsURL, nil)
	if err != nil {
		return result, fmt.Errorf("failed to create request: %w", err)
	}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		// Try the chat completions endpoint as a fallback
		chatURL := baseURL + "/v1/chat/completions"
		req, err = http.NewRequestWithContext(ctx, "OPTIONS", chatURL, nil)
		if err != nil {
			return result, fmt.Errorf("endpoint not accessible: %w", err)
		}

		resp, err = f.client.Do(req)
		if err != nil {
			return result, fmt.Errorf("endpoint not accessible: %w", err)
		}
		resp.Body.Close()

		// If OPTIONS works, the endpoint is accessible but models might not be available
		result.Accessible = true
		result.Message = "Endpoint accessible but /v1/models not available"
		return result, nil
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		result.Accessible = true
		result.ModelsAvailable = true

		var modelsResp ModelsAPIResponse
		if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err == nil {
			result.ModelCount = len(modelsResp.Data)
			if len(modelsResp.Data) > 0 {
				names := make([]string, 0, min(5, len(modelsResp.Data)))
				for i, m := range modelsResp.Data {
					if i >= 5 {
						break
					}
					names = append(names, m.ID)
				}
				result.Message = fmt.Sprintf("Found %d models: %s", len(modelsResp.Data), strings.Join(names, ", "))
				if len(modelsResp.Data) > 5 {
					result.Message += fmt.Sprintf(" (and %d more)", len(modelsResp.Data)-5)
				}
			}
		}

	case http.StatusUnauthorized, http.StatusForbidden:
		result.Accessible = true
		result.AuthRequired = true
		result.Message = "Authentication required"

	case http.StatusNotFound:
		result.Accessible = true
		result.Message = "Models endpoint not available"

	default:
		body, _ := io.ReadAll(resp.Body)
		result.Message = fmt.Sprintf("Unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return result, nil
}

// EndpointTestResult holds the result of an endpoint test
type EndpointTestResult struct {
	Accessible      bool   `json:"accessible"`
	ModelsAvailable bool   `json:"models_available"`
	AuthRequired    bool   `json:"auth_required"`
	ModelCount      int    `json:"model_count"`
	Message         string `json:"message"`
}

// ModelsAPIResponse represents the /v1/models API response
type ModelsAPIResponse struct {
	Object string          `json:"object"`
	Data   []APIModelInfo  `json:"data"`
}

// APIModelInfo represents a model from the API
type APIModelInfo struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Created    int64  `json:"created"`
	OwnedBy    string `json:"owned_by"`
	Permission []struct {
		AllowCreateEngine  bool   `json:"allow_create_engine"`
		AllowSampling      bool   `json:"allow_sampling"`
		AllowLogprobs      bool   `json:"allow_logprobs"`
		AllowSearchIndices bool   `json:"allow_search_indices"`
		AllowView          bool   `json:"allow_view"`
		AllowFineTuning    bool   `json:"allow_fine_tuning"`
		Organization       string `json:"organization"`
		IsBlocking         bool   `json:"is_blocking"`
	} `json:"permission,omitempty"`
	Root   string `json:"root,omitempty"`
	Parent string `json:"parent,omitempty"`
}

// convertAPIModel converts an API model to our internal model format
func convertAPIModel(m APIModelInfo) llm.Model {
	return llm.Model{
		ID:             m.ID,
		Name:           formatModelName(m.ID),
		Description:    fmt.Sprintf("Custom model: %s", m.ID),
		ContextWindow:  estimateContextWindow(m.ID),
		SupportsTools:  detectToolSupport(m.ID),
		SupportsVision: detectVisionSupport(m.ID),
	}
}

// formatModelName creates a user-friendly name from a model ID
func formatModelName(id string) string {
	// Convert dashes and underscores to spaces
	name := strings.ReplaceAll(id, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")

	// Capitalize first letter of each word (simple title case)
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}

	return strings.Join(words, " ")
}

// estimateContextWindow estimates the context window based on model name
func estimateContextWindow(modelID string) int {
	modelLower := strings.ToLower(modelID)

	// Check for explicit context size in name
	contextPatterns := map[string]int{
		"1m":   1000000,
		"128k": 128000,
		"100k": 100000,
		"64k":  64000,
		"32k":  32000,
		"16k":  16000,
		"8k":   8000,
		"4k":   4096,
		"2k":   2048,
	}

	for pattern, size := range contextPatterns {
		if strings.Contains(modelLower, pattern) {
			return size
		}
	}

	// Known model family defaults
	modelFamilies := map[string]int{
		"llama-3.1":   128000,
		"llama-3.2":   128000,
		"llama3.1":    128000,
		"llama3.2":    128000,
		"llama-3":     8192,
		"llama3":      8192,
		"llama-2":     4096,
		"llama2":      4096,
		"mistral":     32000,
		"mixtral":     32000,
		"qwen2":       32000,
		"qwen":        8192,
		"phi-3":       128000,
		"phi-2":       2048,
		"phi":         2048,
		"gemma-2":     8192,
		"gemma":       8192,
		"codellama":   16384,
		"deepseek":    32000,
		"yi":          32000,
		"command-r":   128000,
		"orca":        4096,
		"vicuna":      4096,
		"falcon":      2048,
		"starcoder":   8192,
		"codestral":   32000,
	}

	for family, size := range modelFamilies {
		if strings.Contains(modelLower, family) {
			return size
		}
	}

	// Default
	return 4096
}

// detectToolSupport estimates if a model supports tool calling based on its name
func detectToolSupport(modelID string) bool {
	modelLower := strings.ToLower(modelID)

	// Models known to support tools
	toolCapable := []string{
		"llama-3.1",
		"llama3.1",
		"llama-3.2",
		"llama3.2",
		"mistral",
		"mixtral",
		"qwen",
		"phi-3",
		"gemma-2",
		"command-r",
		"deepseek",
		"yi",
		"codestral",
		"instruct",
		"chat",
	}

	for _, pattern := range toolCapable {
		if strings.Contains(modelLower, pattern) {
			return true
		}
	}

	return false
}

// detectVisionSupport estimates if a model supports vision based on its name
func detectVisionSupport(modelID string) bool {
	modelLower := strings.ToLower(modelID)

	// Models known to support vision
	visionCapable := []string{
		"vision",
		"llava",
		"bakllava",
		"llama-3.2",
		"llama3.2",
		"phi-3.5-vision",
		"moondream",
		"cogvlm",
		"internvl",
	}

	for _, pattern := range visionCapable {
		if strings.Contains(modelLower, pattern) {
			return true
		}
	}

	return false
}

// ParseModelsJSON parses a JSON string containing model definitions
func ParseModelsJSON(jsonStr string) ([]llm.Model, error) {
	var models []llm.Model
	if err := json.Unmarshal([]byte(jsonStr), &models); err != nil {
		// Try parsing as array of model IDs
		var modelIDs []string
		if err := json.Unmarshal([]byte(jsonStr), &modelIDs); err != nil {
			return nil, fmt.Errorf("failed to parse models: %w", err)
		}

		models = make([]llm.Model, len(modelIDs))
		for i, id := range modelIDs {
			models[i] = llm.Model{
				ID:            id,
				Name:          formatModelName(id),
				ContextWindow: estimateContextWindow(id),
				SupportsTools: detectToolSupport(id),
				SupportsVision: detectVisionSupport(id),
			}
		}
	}

	return models, nil
}

// SerializeModels converts models to a JSON string for storage
func SerializeModels(models []llm.Model) (string, error) {
	data, err := json.Marshal(models)
	if err != nil {
		return "", fmt.Errorf("failed to serialize models: %w", err)
	}
	return string(data), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
