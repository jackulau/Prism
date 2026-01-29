package perplexity

import "github.com/jacklau/prism/internal/llm"

// models defines available Perplexity Sonar models
var models = []llm.Model{
	{
		ID:             "llama-3.1-sonar-small-128k-online",
		Name:           "Sonar Small (8B)",
		Description:    "Fast, efficient search-augmented model with web search",
		ContextWindow:  128000,
		SupportsTools:  false,
		SupportsVision: false,
		Capabilities:   []string{"search", "citations"},
	},
	{
		ID:             "llama-3.1-sonar-large-128k-online",
		Name:           "Sonar Large (70B)",
		Description:    "Powerful search-augmented model with web search",
		ContextWindow:  128000,
		SupportsTools:  false,
		SupportsVision: false,
		Capabilities:   []string{"search", "citations"},
	},
	{
		ID:             "llama-3.1-sonar-huge-128k-online",
		Name:           "Sonar Huge (405B)",
		Description:    "Most capable search-augmented model with comprehensive web search",
		ContextWindow:  128000,
		SupportsTools:  false,
		SupportsVision: false,
		Capabilities:   []string{"search", "citations"},
	},
}
