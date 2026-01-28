package prism

import (
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/providers"
)

// ModelPricing contains pricing information for a specific model
// Prices are per 1 million tokens
type ModelPricing struct {
	InputPricePerMillion  float64
	OutputPricePerMillion float64
}

// modelPrices contains the pricing table for supported models
// Prices are in USD per 1 million tokens
var modelPrices = map[string]map[string]ModelPricing{
	"anthropic": {
		"claude-opus-4-5-20251101": {
			InputPricePerMillion:  15.0,
			OutputPricePerMillion: 75.0,
		},
		"claude-sonnet-4-5-20250929": {
			InputPricePerMillion:  3.0,
			OutputPricePerMillion: 15.0,
		},
		"claude-haiku-4-5-20251001": {
			InputPricePerMillion:  0.80,
			OutputPricePerMillion: 4.0,
		},
		// Legacy models
		"claude-3-opus-20240229": {
			InputPricePerMillion:  15.0,
			OutputPricePerMillion: 75.0,
		},
		"claude-3-sonnet-20240229": {
			InputPricePerMillion:  3.0,
			OutputPricePerMillion: 15.0,
		},
		"claude-3-haiku-20240307": {
			InputPricePerMillion:  0.25,
			OutputPricePerMillion: 1.25,
		},
	},
	"openai": {
		"gpt-4o": {
			InputPricePerMillion:  2.50,
			OutputPricePerMillion: 10.0,
		},
		"gpt-4o-mini": {
			InputPricePerMillion:  0.15,
			OutputPricePerMillion: 0.60,
		},
		"gpt-4-turbo": {
			InputPricePerMillion:  10.0,
			OutputPricePerMillion: 30.0,
		},
		"gpt-4": {
			InputPricePerMillion:  30.0,
			OutputPricePerMillion: 60.0,
		},
		"gpt-3.5-turbo": {
			InputPricePerMillion:  0.50,
			OutputPricePerMillion: 1.50,
		},
		"o1-preview": {
			InputPricePerMillion:  15.0,
			OutputPricePerMillion: 60.0,
		},
		"o1-mini": {
			InputPricePerMillion:  3.0,
			OutputPricePerMillion: 12.0,
		},
	},
	"google": {
		"gemini-2.0-flash": {
			InputPricePerMillion:  0.10,
			OutputPricePerMillion: 0.40,
		},
		"gemini-1.5-pro": {
			InputPricePerMillion:  1.25,
			OutputPricePerMillion: 5.0,
		},
		"gemini-1.5-flash": {
			InputPricePerMillion:  0.075,
			OutputPricePerMillion: 0.30,
		},
	},
	"ollama": {
		// Ollama is self-hosted, so no per-token costs
		"*": {
			InputPricePerMillion:  0.0,
			OutputPricePerMillion: 0.0,
		},
	},
}

// GetModelPricing returns the pricing for a specific provider and model
// Returns nil if the model is not found
func GetModelPricing(provider, model string) *ModelPricing {
	providerPrices, ok := modelPrices[provider]
	if !ok {
		return nil
	}

	// Check for exact match
	if pricing, ok := providerPrices[model]; ok {
		return &pricing
	}

	// Check for wildcard (for providers like Ollama)
	if pricing, ok := providerPrices["*"]; ok {
		return &pricing
	}

	return nil
}

// CalculateCost calculates the cost based on token usage and model pricing
func CalculateCost(provider, model string, usage *llm.Usage) *providers.Cost {
	if usage == nil {
		return nil
	}

	pricing := GetModelPricing(provider, model)
	if pricing == nil {
		// Unknown model, return zero cost
		return &providers.Cost{
			InputCost:  0,
			OutputCost: 0,
			TotalCost:  0,
			Currency:   "USD",
		}
	}

	inputCost := float64(usage.PromptTokens) * pricing.InputPricePerMillion / 1_000_000
	outputCost := float64(usage.CompletionTokens) * pricing.OutputPricePerMillion / 1_000_000

	return &providers.Cost{
		InputCost:  inputCost,
		OutputCost: outputCost,
		TotalCost:  inputCost + outputCost,
		Currency:   "USD",
	}
}

// CalculateCostFromUsage calculates cost from a providers.Usage struct
func CalculateCostFromUsage(provider, model string, usage *providers.Usage) *providers.Cost {
	if usage == nil {
		return nil
	}

	pricing := GetModelPricing(provider, model)
	if pricing == nil {
		return &providers.Cost{
			InputCost:  0,
			OutputCost: 0,
			TotalCost:  0,
			Currency:   "USD",
		}
	}

	inputCost := float64(usage.PromptTokens) * pricing.InputPricePerMillion / 1_000_000
	outputCost := float64(usage.CompletionTokens) * pricing.OutputPricePerMillion / 1_000_000

	return &providers.Cost{
		InputCost:  inputCost,
		OutputCost: outputCost,
		TotalCost:  inputCost + outputCost,
		Currency:   "USD",
	}
}

// EstimateCost estimates the cost for a given number of tokens
func EstimateCost(provider, model string, inputTokens, outputTokens int) *providers.Cost {
	pricing := GetModelPricing(provider, model)
	if pricing == nil {
		return &providers.Cost{
			InputCost:  0,
			OutputCost: 0,
			TotalCost:  0,
			Currency:   "USD",
		}
	}

	inputCost := float64(inputTokens) * pricing.InputPricePerMillion / 1_000_000
	outputCost := float64(outputTokens) * pricing.OutputPricePerMillion / 1_000_000

	return &providers.Cost{
		InputCost:  inputCost,
		OutputCost: outputCost,
		TotalCost:  inputCost + outputCost,
		Currency:   "USD",
	}
}

// SupportedProviders returns a list of providers with known pricing
func SupportedProviders() []string {
	providers := make([]string, 0, len(modelPrices))
	for provider := range modelPrices {
		providers = append(providers, provider)
	}
	return providers
}

// SupportedModels returns a list of models with known pricing for a provider
func SupportedModels(provider string) []string {
	providerPrices, ok := modelPrices[provider]
	if !ok {
		return nil
	}

	models := make([]string, 0, len(providerPrices))
	for model := range providerPrices {
		if model != "*" { // Exclude wildcard
			models = append(models, model)
		}
	}
	return models
}
