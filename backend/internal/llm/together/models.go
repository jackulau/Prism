package together

import (
	"github.com/jacklau/prism/internal/llm"
)

// GetModels returns the list of available Together AI models
func GetModels() []llm.Model {
	return []llm.Model{
		// Meta Llama 3.3 Models
		{
			ID:             "meta-llama/Llama-3.3-70B-Instruct-Turbo",
			Name:           "Llama 3.3 70B Instruct Turbo",
			Description:    "Latest Llama model with improved instruction following",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Meta Llama 3.2 Models
		{
			ID:             "meta-llama/Llama-3.2-90B-Vision-Instruct-Turbo",
			Name:           "Llama 3.2 90B Vision Instruct Turbo",
			Description:    "Multimodal model with vision capabilities",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "meta-llama/Llama-3.2-11B-Vision-Instruct-Turbo",
			Name:           "Llama 3.2 11B Vision Instruct Turbo",
			Description:    "Efficient multimodal model with vision",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "meta-llama/Llama-3.2-3B-Instruct-Turbo",
			Name:           "Llama 3.2 3B Instruct Turbo",
			Description:    "Fast, lightweight Llama model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Meta Llama 3.1 Models
		{
			ID:             "meta-llama/Meta-Llama-3.1-405B-Instruct-Turbo",
			Name:           "Llama 3.1 405B Instruct Turbo",
			Description:    "Largest open source model available",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo",
			Name:           "Llama 3.1 70B Instruct Turbo",
			Description:    "High performance Llama 3.1 model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo",
			Name:           "Llama 3.1 8B Instruct Turbo",
			Description:    "Fast and efficient Llama 3.1 model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Qwen 2.5 Models
		{
			ID:             "Qwen/Qwen2.5-72B-Instruct-Turbo",
			Name:           "Qwen 2.5 72B Instruct Turbo",
			Description:    "Powerful Qwen model with strong reasoning",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "Qwen/Qwen2.5-7B-Instruct-Turbo",
			Name:           "Qwen 2.5 7B Instruct Turbo",
			Description:    "Efficient Qwen model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "Qwen/Qwen2.5-Coder-32B-Instruct",
			Name:           "Qwen 2.5 Coder 32B Instruct",
			Description:    "Specialized coding model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "Qwen/QwQ-32B-Preview",
			Name:           "QwQ 32B Preview",
			Description:    "Qwen reasoning model with chain-of-thought",
			ContextWindow:  32768,
			SupportsTools:  false,
			SupportsVision: false,
		},

		// DeepSeek Models
		{
			ID:             "deepseek-ai/DeepSeek-R1",
			Name:           "DeepSeek R1",
			Description:    "Advanced reasoning model",
			ContextWindow:  65536,
			SupportsTools:  false,
			SupportsVision: false,
		},
		{
			ID:             "deepseek-ai/DeepSeek-R1-Distill-Llama-70B",
			Name:           "DeepSeek R1 Distill Llama 70B",
			Description:    "Distilled reasoning model on Llama architecture",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "deepseek-ai/DeepSeek-V3",
			Name:           "DeepSeek V3",
			Description:    "Latest DeepSeek general model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Mistral Models
		{
			ID:             "mistralai/Mixtral-8x22B-Instruct-v0.1",
			Name:           "Mixtral 8x22B Instruct",
			Description:    "Large mixture of experts model",
			ContextWindow:  65536,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "mistralai/Mixtral-8x7B-Instruct-v0.1",
			Name:           "Mixtral 8x7B Instruct",
			Description:    "Efficient mixture of experts model",
			ContextWindow:  32768,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "mistralai/Mistral-7B-Instruct-v0.3",
			Name:           "Mistral 7B Instruct v0.3",
			Description:    "Fast and capable base model",
			ContextWindow:  32768,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Google Gemma Models
		{
			ID:             "google/gemma-2-27b-it",
			Name:           "Gemma 2 27B Instruct",
			Description:    "Google's open Gemma model",
			ContextWindow:  8192,
			SupportsTools:  false,
			SupportsVision: false,
		},
		{
			ID:             "google/gemma-2-9b-it",
			Name:           "Gemma 2 9B Instruct",
			Description:    "Efficient Gemma model",
			ContextWindow:  8192,
			SupportsTools:  false,
			SupportsVision: false,
		},

		// DBRX Model
		{
			ID:             "databricks/dbrx-instruct",
			Name:           "DBRX Instruct",
			Description:    "Databricks MoE model",
			ContextWindow:  32768,
			SupportsTools:  true,
			SupportsVision: false,
		},
	}
}
