package fireworks

import (
	"github.com/jacklau/prism/internal/llm"
)

// DefaultModels returns the list of available Fireworks models
func DefaultModels() []llm.Model {
	return []llm.Model{
		// FireFunction - Specialized function calling models
		{
			ID:             "accounts/fireworks/models/firefunction-v2",
			Name:           "FireFunction v2",
			Description:    "Specialized function calling model with excellent reliability",
			ContextWindow:  8192,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Llama 3.3
		{
			ID:             "accounts/fireworks/models/llama-v3p3-70b-instruct",
			Name:           "Llama 3.3 70B Instruct",
			Description:    "Latest Llama with strong instruction following",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Llama 3.2 - Vision capable
		{
			ID:             "accounts/fireworks/models/llama-v3p2-90b-vision-instruct",
			Name:           "Llama 3.2 90B Vision",
			Description:    "Multimodal model with vision capabilities",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "accounts/fireworks/models/llama-v3p2-11b-vision-instruct",
			Name:           "Llama 3.2 11B Vision",
			Description:    "Compact multimodal model with vision",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "accounts/fireworks/models/llama-v3p2-3b-instruct",
			Name:           "Llama 3.2 3B Instruct",
			Description:    "Fast lightweight model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "accounts/fireworks/models/llama-v3p2-1b-instruct",
			Name:           "Llama 3.2 1B Instruct",
			Description:    "Ultra-fast minimal model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Llama 3.1
		{
			ID:             "accounts/fireworks/models/llama-v3p1-405b-instruct",
			Name:           "Llama 3.1 405B Instruct",
			Description:    "Largest Llama model with best quality",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "accounts/fireworks/models/llama-v3p1-70b-instruct",
			Name:           "Llama 3.1 70B Instruct",
			Description:    "High quality instruction following",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "accounts/fireworks/models/llama-v3p1-8b-instruct",
			Name:           "Llama 3.1 8B Instruct",
			Description:    "Fast and efficient model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Mixtral
		{
			ID:             "accounts/fireworks/models/mixtral-8x22b-instruct",
			Name:           "Mixtral 8x22B Instruct",
			Description:    "Large MoE model with excellent quality",
			ContextWindow:  65536,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "accounts/fireworks/models/mixtral-8x7b-instruct",
			Name:           "Mixtral 8x7B Instruct",
			Description:    "Efficient MoE model",
			ContextWindow:  32768,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Qwen 2.5
		{
			ID:             "accounts/fireworks/models/qwen2p5-72b-instruct",
			Name:           "Qwen 2.5 72B Instruct",
			Description:    "Powerful multilingual model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "accounts/fireworks/models/qwen2p5-coder-32b-instruct",
			Name:           "Qwen 2.5 Coder 32B",
			Description:    "Specialized coding model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Yi
		{
			ID:             "accounts/fireworks/models/yi-large",
			Name:           "Yi Large",
			Description:    "High quality general purpose model",
			ContextWindow:  32768,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// DeepSeek
		{
			ID:             "accounts/fireworks/models/deepseek-v3",
			Name:           "DeepSeek V3",
			Description:    "Advanced reasoning and coding model",
			ContextWindow:  65536,
			SupportsTools:  true,
			SupportsVision: false,
		},
	}
}
