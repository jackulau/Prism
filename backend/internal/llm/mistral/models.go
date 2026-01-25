package mistral

import "github.com/jacklau/prism/internal/llm"

// GetModels returns all available Mistral models
func GetModels() []llm.Model {
	return []llm.Model{
		// Flagship models
		{
			ID:             "mistral-large-latest",
			Name:           "Mistral Large 2",
			Description:    "Flagship model with top-tier reasoning and multilingual capabilities",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: false,
			Capabilities:   []string{"reasoning", "multilingual", "coding"},
		},
		// Efficient models
		{
			ID:             "mistral-small-latest",
			Name:           "Mistral Small",
			Description:    "Cost-efficient model for simple tasks",
			ContextWindow:  32000,
			SupportsTools:  true,
			SupportsVision: false,
			Capabilities:   []string{"fast", "efficient"},
		},
		// Open weights models
		{
			ID:             "open-mistral-nemo",
			Name:           "Mistral NeMo",
			Description:    "Open weights 12B model with 128k context",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: false,
			Capabilities:   []string{"open-weights", "multilingual"},
		},
		// Code-specialized models
		{
			ID:             "codestral-latest",
			Name:           "Codestral",
			Description:    "Specialized model for code generation and completion",
			ContextWindow:  32000,
			SupportsTools:  true,
			SupportsVision: false,
			Capabilities:   []string{"coding", "fill-in-middle"},
		},
		// Vision models
		{
			ID:             "pixtral-large-latest",
			Name:           "Pixtral Large",
			Description:    "Multimodal model with vision capabilities",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: true,
			Capabilities:   []string{"vision", "reasoning", "multilingual"},
		},
		{
			ID:             "pixtral-12b-latest",
			Name:           "Pixtral 12B",
			Description:    "Efficient multimodal model with vision",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: true,
			Capabilities:   []string{"vision", "efficient"},
		},
		// Edge models
		{
			ID:             "ministral-8b-latest",
			Name:           "Ministral 8B",
			Description:    "Optimized for edge computing and on-device use",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: false,
			Capabilities:   []string{"edge", "fast", "efficient"},
		},
		{
			ID:             "ministral-3b-latest",
			Name:           "Ministral 3B",
			Description:    "Ultra-compact model for edge devices",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: false,
			Capabilities:   []string{"edge", "ultra-fast", "compact"},
		},
	}
}
