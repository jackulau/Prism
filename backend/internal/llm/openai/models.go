package openai

// ModelConfig contains configuration and capabilities for a model
type ModelConfig struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	ContextWindow  int     `json:"context_window"`
	MaxOutput      int     `json:"max_output"`
	SupportsVision bool    `json:"supports_vision"`
	SupportsTools  bool    `json:"supports_tools"`
	SupportsJSON   bool    `json:"supports_json"`
	InputCostPer1M float64 `json:"input_cost_per_1m"`  // Cost per 1M input tokens
	OutputCostPer1M float64 `json:"output_cost_per_1m"` // Cost per 1M output tokens
}

// ModelConfigs contains configurations for all supported models
var ModelConfigs = map[string]ModelConfig{
	"o3": {
		ID:             "o3",
		Name:           "OpenAI o3",
		ContextWindow:  200000,
		MaxOutput:      100000,
		SupportsVision: true,
		SupportsTools:  true,
		SupportsJSON:   true,
		InputCostPer1M:  10.0,
		OutputCostPer1M: 40.0,
	},
	"o4-mini": {
		ID:             "o4-mini",
		Name:           "OpenAI o4-mini",
		ContextWindow:  128000,
		MaxOutput:      65536,
		SupportsVision: true,
		SupportsTools:  true,
		SupportsJSON:   true,
		InputCostPer1M:  1.10,
		OutputCostPer1M: 4.40,
	},
	"gpt-4.1": {
		ID:             "gpt-4.1",
		Name:           "GPT-4.1",
		ContextWindow:  1000000,
		MaxOutput:      32768,
		SupportsVision: true,
		SupportsTools:  true,
		SupportsJSON:   true,
		InputCostPer1M:  2.0,
		OutputCostPer1M: 8.0,
	},
	"gpt-4.1-mini": {
		ID:             "gpt-4.1-mini",
		Name:           "GPT-4.1 Mini",
		ContextWindow:  1000000,
		MaxOutput:      32768,
		SupportsVision: true,
		SupportsTools:  true,
		SupportsJSON:   true,
		InputCostPer1M:  0.40,
		OutputCostPer1M: 1.60,
	},
	"gpt-4o": {
		ID:             "gpt-4o",
		Name:           "GPT-4o",
		ContextWindow:  128000,
		MaxOutput:      16384,
		SupportsVision: true,
		SupportsTools:  true,
		SupportsJSON:   true,
		InputCostPer1M:  2.50,
		OutputCostPer1M: 10.0,
	},
	"gpt-4o-mini": {
		ID:             "gpt-4o-mini",
		Name:           "GPT-4o Mini",
		ContextWindow:  128000,
		MaxOutput:      16384,
		SupportsVision: true,
		SupportsTools:  true,
		SupportsJSON:   true,
		InputCostPer1M:  0.15,
		OutputCostPer1M: 0.60,
	},
}

// GetModelConfig returns the configuration for a model, or a default if not found
func GetModelConfig(modelID string) ModelConfig {
	if config, ok := ModelConfigs[modelID]; ok {
		return config
	}
	// Return a reasonable default for unknown models
	return ModelConfig{
		ID:             modelID,
		Name:           modelID,
		ContextWindow:  128000,
		MaxOutput:      16384,
		SupportsVision: true,
		SupportsTools:  true,
		SupportsJSON:   true,
		InputCostPer1M:  2.50,
		OutputCostPer1M: 10.0,
	}
}

// EstimateMaxOutputTokens estimates the maximum output tokens based on input tokens
func EstimateMaxOutputTokens(modelID string, inputTokens int) int {
	config := GetModelConfig(modelID)
	remaining := config.ContextWindow - inputTokens
	if remaining <= 0 {
		return 0
	}
	if remaining > config.MaxOutput {
		return config.MaxOutput
	}
	return remaining
}
