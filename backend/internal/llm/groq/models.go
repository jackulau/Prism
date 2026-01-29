package groq

import "github.com/jacklau/prism/internal/llm"

// GetModels returns the list of available Groq models
func GetModels() []llm.Model {
	return []llm.Model{
		// Llama 3.3 models
		{
			ID:             "llama-3.3-70b-versatile",
			Name:           "Llama 3.3 70B Versatile",
			Description:    "Latest Llama model with excellent reasoning and coding capabilities",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "llama-3.3-70b-specdec",
			Name:           "Llama 3.3 70B SpecDec",
			Description:    "Llama 3.3 70B with speculative decoding for faster inference",
			ContextWindow:  8192,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Llama 3.2 Vision models
		{
			ID:             "llama-3.2-90b-vision-preview",
			Name:           "Llama 3.2 90B Vision",
			Description:    "Largest Llama vision model for complex image understanding",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "llama-3.2-11b-vision-preview",
			Name:           "Llama 3.2 11B Vision",
			Description:    "Efficient vision model for image understanding tasks",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: true,
		},

		// Llama 3.2 Text models
		{
			ID:             "llama-3.2-3b-preview",
			Name:           "Llama 3.2 3B",
			Description:    "Fast, lightweight model for simple tasks",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "llama-3.2-1b-preview",
			Name:           "Llama 3.2 1B",
			Description:    "Ultra-fast model for basic text tasks",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Llama 3.1 models
		{
			ID:             "llama-3.1-70b-versatile",
			Name:           "Llama 3.1 70B Versatile",
			Description:    "Powerful model for complex reasoning and coding",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "llama-3.1-8b-instant",
			Name:           "Llama 3.1 8B Instant",
			Description:    "Fast and efficient for quick responses",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Llama Guard (safety model)
		{
			ID:             "llama-guard-3-8b",
			Name:           "Llama Guard 3 8B",
			Description:    "Safety classification model for content moderation",
			ContextWindow:  8192,
			SupportsTools:  false,
			SupportsVision: false,
		},

		// Mixtral models
		{
			ID:             "mixtral-8x7b-32768",
			Name:           "Mixtral 8x7B",
			Description:    "Mixture of experts model with strong performance",
			ContextWindow:  32768,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Gemma models
		{
			ID:             "gemma2-9b-it",
			Name:           "Gemma 2 9B",
			Description:    "Google's efficient instruction-tuned model",
			ContextWindow:  8192,
			SupportsTools:  true,
			SupportsVision: false,
		},
	}
}
