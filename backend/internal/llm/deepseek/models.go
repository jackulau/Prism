package deepseek

import "github.com/jacklau/prism/internal/llm"

// GetModels returns the list of available DeepSeek models
func GetModels() []llm.Model {
	return []llm.Model{
		{
			ID:            "deepseek-chat",
			Name:          "DeepSeek V3",
			Description:   "Flagship model with excellent coding and reasoning",
			ContextWindow: 64000,
			SupportsTools: true,
			SupportsVision: false,
			Capabilities:  []string{"chat", "coding", "reasoning"},
		},
		{
			ID:            "deepseek-reasoner",
			Name:          "DeepSeek R1",
			Description:   "Advanced reasoning model - thinks before answering",
			ContextWindow: 64000,
			SupportsTools: true,
			SupportsVision: false,
			Capabilities:  []string{"chat", "reasoning", "chain-of-thought"},
		},
		{
			ID:            "deepseek-coder",
			Name:          "DeepSeek Coder",
			Description:   "Specialized model for code generation and understanding",
			ContextWindow: 64000,
			SupportsTools: true,
			SupportsVision: false,
			Capabilities:  []string{"coding", "code-completion", "code-explanation"},
		},
	}
}
