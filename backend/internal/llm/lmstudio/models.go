package lmstudio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jacklau/prism/internal/llm"
)

// modelsResponse represents the response from LM Studio's /v1/models endpoint
type modelsResponse struct {
	Object string `json:"object"`
	Data   []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

// fetchModels fetches the currently loaded models from LM Studio
func (c *Client) fetchModels() ([]llm.Model, error) {
	resp, err := c.client.Get(c.baseURL + "/v1/models")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LM Studio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LM Studio not available: status %d", resp.StatusCode)
	}

	var result modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	models := make([]llm.Model, len(result.Data))
	for i, m := range result.Data {
		// Detect model capabilities based on model name
		supportsTools := isToolCapableModel(m.ID)
		supportsVision := isVisionCapableModel(m.ID)
		contextWindow := getModelContextWindow(m.ID)

		// Create a readable name from the model ID
		name := formatModelName(m.ID)

		models[i] = llm.Model{
			ID:             m.ID,
			Name:           name,
			Description:    "Local model via LM Studio",
			ContextWindow:  contextWindow,
			SupportsTools:  supportsTools,
			SupportsVision: supportsVision,
		}
	}

	return models, nil
}

// formatModelName creates a human-readable name from the model ID
func formatModelName(modelID string) string {
	// Remove common path prefixes and extensions
	name := modelID

	// Remove file extensions
	name = strings.TrimSuffix(name, ".gguf")
	name = strings.TrimSuffix(name, ".GGUF")

	// Replace underscores and hyphens with spaces for readability
	// but keep the original if it's a common model name format
	if strings.Contains(name, "/") {
		// Extract just the model name from path-like IDs
		parts := strings.Split(name, "/")
		name = parts[len(parts)-1]
	}

	return name
}

// isToolCapableModel checks if a model supports tool calling
func isToolCapableModel(modelName string) bool {
	toolCapableModels := []string{
		// Meta Llama models (3.1+ support tools)
		"llama3.1", "llama3.2", "llama3.3", "llama-3.1", "llama-3.2", "llama-3.3",
		"llama-3.1", "llama-3.2", "llama-3.3",

		// Mistral AI models
		"mistral", "mixtral", "codestral", "mistral-nemo", "mistral-small", "mistral-large",

		// Alibaba Qwen models
		"qwen2", "qwen2.5", "qwen-2", "qwen-2.5", "qwq",

		// Cohere Command models
		"command-r", "command-r-plus", "c4ai",

		// Function-calling specialized models
		"firefunction", "functionary", "gorilla", "nexusraven",

		// Nous Research models
		"hermes", "nous-hermes", "openhermes",

		// IBM Granite models
		"granite",

		// DeepSeek models
		"deepseek", "deepseek-coder", "deepseek-v2", "deepseek-v3",

		// Microsoft Phi models
		"phi-3", "phi-4", "phi3", "phi4",

		// Google Gemma models (2.0+ with tool support)
		"gemma2", "gemma-2",

		// 01.ai Yi models
		"yi-", "yi1.5", "yi-1.5",

		// Other tool-capable models
		"internlm", "glm", "chatglm",
		"solar", "solar-pro",
		"dolphin", "dolphin-mistral", "dolphin-llama",
		"nemotron",
		"smollm2",
		"athene",
		"marco",
	}

	nameLower := strings.ToLower(modelName)
	for _, tcm := range toolCapableModels {
		if strings.Contains(nameLower, tcm) {
			return true
		}
	}
	return false
}

// isVisionCapableModel checks if a model supports vision/image input
func isVisionCapableModel(modelName string) bool {
	visionModels := []string{
		// LLaVA models
		"llava", "llava-llama3", "llava-phi3",
		// Generic vision indicators
		"vision", "-v", "vl",
		// Specific vision models
		"bakllava", "moondream",
		"minicpm-v", "internvl",
		"cogvlm", "qwen-vl", "qwen2-vl",
		"llama3.2-vision", "llama-3.2-vision",
		"pixtral",
	}

	nameLower := strings.ToLower(modelName)
	for _, vm := range visionModels {
		if strings.Contains(nameLower, vm) {
			return true
		}
	}
	return false
}

// getModelContextWindow returns the estimated context window for a model
func getModelContextWindow(modelName string) int {
	nameLower := strings.ToLower(modelName)

	// Models with 128k+ context
	if strings.Contains(nameLower, "llama3.1") || strings.Contains(nameLower, "llama-3.1") ||
		strings.Contains(nameLower, "llama3.2") || strings.Contains(nameLower, "llama-3.2") ||
		strings.Contains(nameLower, "llama3.3") || strings.Contains(nameLower, "llama-3.3") {
		return 131072 // 128k
	}
	if strings.Contains(nameLower, "qwen2.5") || strings.Contains(nameLower, "qwen-2.5") ||
		strings.Contains(nameLower, "qwq") {
		return 131072 // 128k
	}
	if strings.Contains(nameLower, "deepseek-v3") || strings.Contains(nameLower, "deepseek-v2") {
		return 131072 // 128k
	}
	if strings.Contains(nameLower, "mistral-large") || strings.Contains(nameLower, "mistral-nemo") {
		return 131072 // 128k
	}
	if strings.Contains(nameLower, "gemma2") || strings.Contains(nameLower, "gemma-2") {
		return 8192
	}

	// Models with 32k context
	if strings.Contains(nameLower, "mistral") || strings.Contains(nameLower, "mixtral") {
		return 32768 // 32k
	}
	if strings.Contains(nameLower, "qwen2") || strings.Contains(nameLower, "qwen-2") {
		return 32768 // 32k
	}
	if strings.Contains(nameLower, "command-r") {
		return 131072 // 128k
	}
	if strings.Contains(nameLower, "yi-") || strings.Contains(nameLower, "yi1.5") {
		return 32768 // 32k
	}
	if strings.Contains(nameLower, "deepseek") {
		return 32768 // 32k
	}

	// Models with 16k context
	if strings.Contains(nameLower, "phi-3") || strings.Contains(nameLower, "phi3") ||
		strings.Contains(nameLower, "phi-4") || strings.Contains(nameLower, "phi4") {
		return 16384 // 16k
	}
	if strings.Contains(nameLower, "granite") {
		return 8192
	}

	// Models with 8k context
	if strings.Contains(nameLower, "llama3") || strings.Contains(nameLower, "llama-3") {
		return 8192 // 8k (base llama3 without extended context)
	}
	if strings.Contains(nameLower, "hermes") || strings.Contains(nameLower, "openhermes") {
		return 8192
	}

	// Default context window for LM Studio models
	return 4096
}
