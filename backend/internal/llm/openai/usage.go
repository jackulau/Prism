package openai

import (
	"sync"

	"github.com/jacklau/prism/internal/llm"
)

// UsageTracker tracks token usage across models
type UsageTracker struct {
	mu    sync.Mutex
	usage map[string]*llm.Usage
}

// NewUsageTracker creates a new usage tracker
func NewUsageTracker() *UsageTracker {
	return &UsageTracker{
		usage: make(map[string]*llm.Usage),
	}
}

// Record records usage for a model
func (ut *UsageTracker) Record(model string, usage *llm.Usage) {
	if usage == nil {
		return
	}

	ut.mu.Lock()
	defer ut.mu.Unlock()

	if ut.usage[model] == nil {
		ut.usage[model] = &llm.Usage{}
	}

	ut.usage[model].PromptTokens += usage.PromptTokens
	ut.usage[model].CompletionTokens += usage.CompletionTokens
	ut.usage[model].TotalTokens += usage.TotalTokens
}

// GetUsage returns usage for a specific model
func (ut *UsageTracker) GetUsage(model string) *llm.Usage {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	if usage, ok := ut.usage[model]; ok {
		// Return a copy to avoid race conditions
		return &llm.Usage{
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			TotalTokens:      usage.TotalTokens,
		}
	}
	return &llm.Usage{}
}

// GetTotalUsage returns total usage across all models
func (ut *UsageTracker) GetTotalUsage() *llm.Usage {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	total := &llm.Usage{}
	for _, usage := range ut.usage {
		total.PromptTokens += usage.PromptTokens
		total.CompletionTokens += usage.CompletionTokens
		total.TotalTokens += usage.TotalTokens
	}
	return total
}

// GetAllUsage returns usage for all models
func (ut *UsageTracker) GetAllUsage() map[string]*llm.Usage {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	result := make(map[string]*llm.Usage, len(ut.usage))
	for model, usage := range ut.usage {
		result[model] = &llm.Usage{
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			TotalTokens:      usage.TotalTokens,
		}
	}
	return result
}

// Reset resets all usage tracking
func (ut *UsageTracker) Reset() {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	ut.usage = make(map[string]*llm.Usage)
}

// ResetModel resets usage for a specific model
func (ut *UsageTracker) ResetModel(model string) {
	ut.mu.Lock()
	defer ut.mu.Unlock()
	delete(ut.usage, model)
}

// EstimateCost estimates the cost based on usage and model pricing
func (ut *UsageTracker) EstimateCost(model string) float64 {
	usage := ut.GetUsage(model)
	config := GetModelConfig(model)

	inputCost := float64(usage.PromptTokens) * config.InputCostPer1M / 1_000_000
	outputCost := float64(usage.CompletionTokens) * config.OutputCostPer1M / 1_000_000

	return inputCost + outputCost
}

// EstimateTotalCost estimates total cost across all models
func (ut *UsageTracker) EstimateTotalCost() float64 {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	var total float64
	for model, usage := range ut.usage {
		config := GetModelConfig(model)
		inputCost := float64(usage.PromptTokens) * config.InputCostPer1M / 1_000_000
		outputCost := float64(usage.CompletionTokens) * config.OutputCostPer1M / 1_000_000
		total += inputCost + outputCost
	}
	return total
}

// UsageSummary provides a summary of usage with costs
type UsageSummary struct {
	Model            string  `json:"model"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCost    float64 `json:"estimated_cost_usd"`
}

// GetSummary returns usage summaries for all models
func (ut *UsageTracker) GetSummary() []UsageSummary {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	summaries := make([]UsageSummary, 0, len(ut.usage))
	for model, usage := range ut.usage {
		config := GetModelConfig(model)
		inputCost := float64(usage.PromptTokens) * config.InputCostPer1M / 1_000_000
		outputCost := float64(usage.CompletionTokens) * config.OutputCostPer1M / 1_000_000

		summaries = append(summaries, UsageSummary{
			Model:            model,
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			TotalTokens:      usage.TotalTokens,
			EstimatedCost:    inputCost + outputCost,
		})
	}
	return summaries
}
