package openrouter

import "github.com/jacklau/prism/internal/llm"

// GetPopularModels returns a curated list of popular models available on OpenRouter.
// This provides default models when the API hasn't been queried yet.
func GetPopularModels() []llm.Model {
	return []llm.Model{
		// Meta Llama models
		{
			ID:             "meta-llama/llama-3.3-70b-instruct",
			Name:           "Llama 3.3 70B Instruct",
			Description:    "Meta's latest Llama model with excellent instruction following",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "meta-llama/llama-3.2-90b-vision-instruct",
			Name:           "Llama 3.2 90B Vision",
			Description:    "Meta's multimodal Llama model with vision capabilities",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "meta-llama/llama-3.1-405b-instruct",
			Name:           "Llama 3.1 405B Instruct",
			Description:    "Meta's largest open model with exceptional capabilities",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "meta-llama/llama-3.1-70b-instruct",
			Name:           "Llama 3.1 70B Instruct",
			Description:    "Excellent balance of capability and speed",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "meta-llama/llama-3.1-8b-instruct",
			Name:           "Llama 3.1 8B Instruct",
			Description:    "Fast and efficient for simpler tasks",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// DeepSeek models
		{
			ID:             "deepseek/deepseek-r1",
			Name:           "DeepSeek R1",
			Description:    "DeepSeek's reasoning model with strong performance",
			ContextWindow:  64000,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "deepseek/deepseek-chat",
			Name:           "DeepSeek Chat",
			Description:    "DeepSeek's conversational model",
			ContextWindow:  64000,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "deepseek/deepseek-coder",
			Name:           "DeepSeek Coder",
			Description:    "Specialized for code generation and understanding",
			ContextWindow:  64000,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Mistral models
		{
			ID:             "mistralai/mistral-large-2411",
			Name:           "Mistral Large (Nov 2024)",
			Description:    "Mistral's flagship model with excellent reasoning",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "mistralai/mistral-medium-2312",
			Name:           "Mistral Medium",
			Description:    "Good balance of performance and cost",
			ContextWindow:  32000,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "mistralai/mixtral-8x7b-instruct",
			Name:           "Mixtral 8x7B Instruct",
			Description:    "Mixture of experts model with high throughput",
			ContextWindow:  32768,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "mistralai/mistral-7b-instruct",
			Name:           "Mistral 7B Instruct",
			Description:    "Fast and efficient smaller model",
			ContextWindow:  32768,
			SupportsTools:  false,
			SupportsVision: false,
		},
		{
			ID:             "mistralai/codestral-2501",
			Name:           "Codestral (Jan 2025)",
			Description:    "Mistral's code-specialized model",
			ContextWindow:  256000,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Qwen models
		{
			ID:             "qwen/qwen-2.5-72b-instruct",
			Name:           "Qwen 2.5 72B Instruct",
			Description:    "Alibaba's powerful open model",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "qwen/qwen-2.5-coder-32b-instruct",
			Name:           "Qwen 2.5 Coder 32B",
			Description:    "Specialized for code generation",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "qwen/qwq-32b-preview",
			Name:           "QwQ 32B Preview",
			Description:    "Qwen's reasoning model",
			ContextWindow:  32768,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Google Gemma models
		{
			ID:             "google/gemma-2-27b-it",
			Name:           "Gemma 2 27B",
			Description:    "Google's efficient open model",
			ContextWindow:  8192,
			SupportsTools:  false,
			SupportsVision: false,
		},
		{
			ID:             "google/gemma-2-9b-it",
			Name:           "Gemma 2 9B",
			Description:    "Fast Google open model",
			ContextWindow:  8192,
			SupportsTools:  false,
			SupportsVision: false,
		},

		// Anthropic models (via OpenRouter)
		{
			ID:             "anthropic/claude-3.5-sonnet",
			Name:           "Claude 3.5 Sonnet",
			Description:    "Anthropic's balanced flagship model",
			ContextWindow:  200000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "anthropic/claude-3-opus",
			Name:           "Claude 3 Opus",
			Description:    "Anthropic's most capable model",
			ContextWindow:  200000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "anthropic/claude-3-haiku",
			Name:           "Claude 3 Haiku",
			Description:    "Fast and affordable Claude model",
			ContextWindow:  200000,
			SupportsTools:  true,
			SupportsVision: true,
		},

		// OpenAI models (via OpenRouter)
		{
			ID:             "openai/gpt-4o",
			Name:           "GPT-4o",
			Description:    "OpenAI's multimodal flagship",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "openai/gpt-4o-mini",
			Name:           "GPT-4o Mini",
			Description:    "Fast and affordable GPT-4 variant",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "openai/gpt-4-turbo",
			Name:           "GPT-4 Turbo",
			Description:    "Enhanced GPT-4 with vision",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: true,
		},

		// Google models (via OpenRouter)
		{
			ID:             "google/gemini-pro-1.5",
			Name:           "Gemini Pro 1.5",
			Description:    "Google's latest Gemini model",
			ContextWindow:  2800000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "google/gemini-flash-1.5",
			Name:           "Gemini Flash 1.5",
			Description:    "Fast Google Gemini variant",
			ContextWindow:  1000000,
			SupportsTools:  true,
			SupportsVision: true,
		},

		// Cohere models
		{
			ID:             "cohere/command-r-plus",
			Name:           "Command R+",
			Description:    "Cohere's flagship model with RAG support",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: false,
		},
		{
			ID:             "cohere/command-r",
			Name:           "Command R",
			Description:    "Fast Cohere model with good reasoning",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// Perplexity models
		{
			ID:             "perplexity/llama-3.1-sonar-huge-128k-online",
			Name:           "Sonar Huge 128K Online",
			Description:    "Real-time web search integrated",
			ContextWindow:  127072,
			SupportsTools:  false,
			SupportsVision: false,
		},
		{
			ID:             "perplexity/llama-3.1-sonar-large-128k-online",
			Name:           "Sonar Large 128K Online",
			Description:    "Web search with good performance",
			ContextWindow:  127072,
			SupportsTools:  false,
			SupportsVision: false,
		},

		// Nous Research models
		{
			ID:             "nousresearch/hermes-3-llama-3.1-405b",
			Name:           "Hermes 3 405B",
			Description:    "Nous Research fine-tuned Llama",
			ContextWindow:  131072,
			SupportsTools:  true,
			SupportsVision: false,
		},

		// 01.ai models
		{
			ID:             "01-ai/yi-large",
			Name:           "Yi Large",
			Description:    "01.ai's powerful model",
			ContextWindow:  32768,
			SupportsTools:  false,
			SupportsVision: false,
		},
	}
}
