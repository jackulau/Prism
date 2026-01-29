package huggingface

import "github.com/jacklau/prism/internal/llm"

// DefaultModels returns the list of popular models available on Hugging Face Hub
// These models support the text-generation-inference format with chat templates
func DefaultModels() []llm.Model {
	return []llm.Model{
		// Meta Llama models
		{
			ID:             "meta-llama/Llama-3.3-70B-Instruct",
			Name:           "Llama 3.3 70B Instruct",
			Description:    "Meta's latest instruction-tuned Llama model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "meta-llama/Llama-3.1-8B-Instruct",
			Name:           "Llama 3.1 8B Instruct",
			Description:    "Fast and efficient Llama model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "meta-llama/Llama-3.2-11B-Vision-Instruct",
			Name:           "Llama 3.2 11B Vision",
			Description:    "Multimodal Llama model with vision support",
			ContextWindow:  131072,
			SupportsTools:  false,
			SupportsVision: true,
		},
		// Mistral models
		{
			ID:             "mistralai/Mistral-7B-Instruct-v0.3",
			Name:           "Mistral 7B Instruct v0.3",
			Description:    "Fast and efficient Mistral model with function calling",
			ContextWindow:  32768,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "mistralai/Mixtral-8x7B-Instruct-v0.1",
			Name:           "Mixtral 8x7B Instruct",
			Description:    "Mixture of experts model with strong performance",
			ContextWindow:  32768,
			SupportsTools:  true,
			SupportsVision: false,
		},
		// Microsoft Phi models
		{
			ID:             "microsoft/Phi-3.5-mini-instruct",
			Name:           "Phi-3.5 Mini Instruct",
			Description:    "Compact but capable Microsoft model",
			ContextWindow:  131072,
			SupportsTools:  false,
			SupportsVision: false,
		},
		// Qwen models
		{
			ID:             "Qwen/Qwen2.5-72B-Instruct",
			Name:           "Qwen 2.5 72B Instruct",
			Description:    "Alibaba's powerful instruction-tuned model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "Qwen/Qwen2.5-Coder-32B-Instruct",
			Name:           "Qwen 2.5 Coder 32B",
			Description:    "Specialized coding model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		// DeepSeek models
		{
			ID:             "deepseek-ai/DeepSeek-R1-Distill-Qwen-32B",
			Name:           "DeepSeek R1 Distill 32B",
			Description:    "Reasoning-focused distilled model",
			ContextWindow:  131072,
			SupportsTools:  false,
			SupportsVision: false,
		},
		// Code models
		{
			ID:             "bigcode/starcoder2-15b-instruct-v0.1",
			Name:           "StarCoder2 15B Instruct",
			Description:    "Specialized code generation model",
			ContextWindow:  16384,
			SupportsTools:  false,
			SupportsVision: false,
		},
	}
}

// ModelSupportsTools returns whether a specific model ID supports tool calling
func ModelSupportsTools(modelID string) bool {
	for _, m := range DefaultModels() {
		if m.ID == modelID {
			return m.SupportsTools
		}
	}
	// For custom models, assume tool support if using OpenAI-compatible endpoint
	return false
}

// ModelSupportsVision returns whether a specific model ID supports vision
func ModelSupportsVision(modelID string) bool {
	for _, m := range DefaultModels() {
		if m.ID == modelID {
			return m.SupportsVision
		}
	}
	return false
}
