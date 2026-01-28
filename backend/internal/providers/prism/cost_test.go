package prism

import (
	"testing"

	"github.com/jacklau/prism/internal/llm"
)

func TestGetModelPricing(t *testing.T) {
	tests := []struct {
		name          string
		provider      string
		model         string
		expectPricing bool
	}{
		{
			name:          "Anthropic Claude Sonnet",
			provider:      "anthropic",
			model:         "claude-sonnet-4-5-20250929",
			expectPricing: true,
		},
		{
			name:          "Anthropic Claude Opus",
			provider:      "anthropic",
			model:         "claude-opus-4-5-20251101",
			expectPricing: true,
		},
		{
			name:          "OpenAI GPT-4o",
			provider:      "openai",
			model:         "gpt-4o",
			expectPricing: true,
		},
		{
			name:          "Google Gemini",
			provider:      "google",
			model:         "gemini-2.0-flash",
			expectPricing: true,
		},
		{
			name:          "Ollama wildcard",
			provider:      "ollama",
			model:         "llama2",
			expectPricing: true,
		},
		{
			name:          "Unknown provider",
			provider:      "unknown",
			model:         "some-model",
			expectPricing: false,
		},
		{
			name:          "Unknown model",
			provider:      "anthropic",
			model:         "unknown-model",
			expectPricing: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pricing := GetModelPricing(tt.provider, tt.model)
			if tt.expectPricing && pricing == nil {
				t.Errorf("expected pricing for %s/%s, got nil", tt.provider, tt.model)
			}
			if !tt.expectPricing && pricing != nil {
				t.Errorf("expected no pricing for %s/%s, got %+v", tt.provider, tt.model, pricing)
			}
		})
	}
}

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name           string
		provider       string
		model          string
		usage          *llm.Usage
		expectedInput  float64
		expectedOutput float64
	}{
		{
			name:     "Claude Sonnet - 1000 tokens each",
			provider: "anthropic",
			model:    "claude-sonnet-4-5-20250929",
			usage: &llm.Usage{
				PromptTokens:     1000,
				CompletionTokens: 1000,
				TotalTokens:      2000,
			},
			expectedInput:  0.003, // 1000 * 3.0 / 1M = 0.003
			expectedOutput: 0.015, // 1000 * 15.0 / 1M = 0.015
		},
		{
			name:     "Claude Opus - 1000000 tokens each (1M)",
			provider: "anthropic",
			model:    "claude-opus-4-5-20251101",
			usage: &llm.Usage{
				PromptTokens:     1000000,
				CompletionTokens: 1000000,
				TotalTokens:      2000000,
			},
			expectedInput:  15.0, // 1M * 15.0 / 1M
			expectedOutput: 75.0, // 1M * 75.0 / 1M
		},
		{
			name:     "Ollama - zero cost",
			provider: "ollama",
			model:    "llama2",
			usage: &llm.Usage{
				PromptTokens:     10000,
				CompletionTokens: 10000,
				TotalTokens:      20000,
			},
			expectedInput:  0.0,
			expectedOutput: 0.0,
		},
		{
			name:           "Nil usage",
			provider:       "anthropic",
			model:          "claude-sonnet-4-5-20250929",
			usage:          nil,
			expectedInput:  0.0,
			expectedOutput: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := CalculateCost(tt.provider, tt.model, tt.usage)

			if tt.usage == nil {
				if cost != nil {
					t.Errorf("expected nil cost for nil usage, got %+v", cost)
				}
				return
			}

			if cost == nil {
				t.Fatal("expected non-nil cost, got nil")
			}

			// Check input cost (with small tolerance for floating point)
			if diff := cost.InputCost - tt.expectedInput; diff > 0.0001 || diff < -0.0001 {
				t.Errorf("input cost = %f, expected %f", cost.InputCost, tt.expectedInput)
			}

			// Check output cost
			if diff := cost.OutputCost - tt.expectedOutput; diff > 0.0001 || diff < -0.0001 {
				t.Errorf("output cost = %f, expected %f", cost.OutputCost, tt.expectedOutput)
			}

			// Check total
			expectedTotal := tt.expectedInput + tt.expectedOutput
			if diff := cost.TotalCost - expectedTotal; diff > 0.0001 || diff < -0.0001 {
				t.Errorf("total cost = %f, expected %f", cost.TotalCost, expectedTotal)
			}

			// Check currency
			if cost.Currency != "USD" {
				t.Errorf("currency = %s, expected USD", cost.Currency)
			}
		})
	}
}

func TestEstimateCost(t *testing.T) {
	cost := EstimateCost("anthropic", "claude-sonnet-4-5-20250929", 1000, 500)

	if cost == nil {
		t.Fatal("expected non-nil cost")
	}

	// 1000 input tokens at $3/1M = $0.003
	expectedInput := 0.003
	// 500 output tokens at $15/1M = $0.0075
	expectedOutput := 0.0075

	if diff := cost.InputCost - expectedInput; diff > 0.0001 || diff < -0.0001 {
		t.Errorf("input cost = %f, expected %f", cost.InputCost, expectedInput)
	}

	if diff := cost.OutputCost - expectedOutput; diff > 0.0001 || diff < -0.0001 {
		t.Errorf("output cost = %f, expected %f", cost.OutputCost, expectedOutput)
	}
}

func TestSupportedProviders(t *testing.T) {
	providers := SupportedProviders()

	if len(providers) == 0 {
		t.Error("expected at least one supported provider")
	}

	// Check that known providers are included
	known := map[string]bool{"anthropic": false, "openai": false, "google": false, "ollama": false}
	for _, p := range providers {
		if _, ok := known[p]; ok {
			known[p] = true
		}
	}

	for provider, found := range known {
		if !found {
			t.Errorf("expected provider %s to be in supported list", provider)
		}
	}
}

func TestSupportedModels(t *testing.T) {
	models := SupportedModels("anthropic")

	if len(models) == 0 {
		t.Error("expected at least one model for anthropic")
	}

	// Check that claude models are included
	foundSonnet := false
	foundOpus := false
	for _, m := range models {
		if m == "claude-sonnet-4-5-20250929" {
			foundSonnet = true
		}
		if m == "claude-opus-4-5-20251101" {
			foundOpus = true
		}
	}

	if !foundSonnet {
		t.Error("expected claude-sonnet-4-5-20250929 in supported models")
	}
	if !foundOpus {
		t.Error("expected claude-opus-4-5-20251101 in supported models")
	}

	// Unknown provider should return nil
	unknownModels := SupportedModels("unknown-provider")
	if unknownModels != nil {
		t.Errorf("expected nil for unknown provider, got %v", unknownModels)
	}
}
