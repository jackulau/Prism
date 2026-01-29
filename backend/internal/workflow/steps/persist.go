package steps

import (
	"context"
	"fmt"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/workflow"
)

// PersistStep handles saving results and tracking token usage
type PersistStep struct {
	messageRepo   *repository.MessageRepository
	tokenUsageRepo *workflow.TokenUsageRepository
	executionRepo  *workflow.AgentExecutionRepository
}

// NewPersistStep creates a new persist step handler
func NewPersistStep(
	messageRepo *repository.MessageRepository,
	tokenUsageRepo *workflow.TokenUsageRepository,
	executionRepo *workflow.AgentExecutionRepository,
) *PersistStep {
	return &PersistStep{
		messageRepo:   messageRepo,
		tokenUsageRepo: tokenUsageRepo,
		executionRepo:  executionRepo,
	}
}

// SaveResponse saves the LLM response and messages to the database (Step 7)
func (s *PersistStep) SaveResponse(ctx context.Context, conversationID string, result *workflow.LLMResult) error {
	if conversationID == "" || result == nil {
		return nil
	}

	// Save messages from the result
	for _, msg := range result.Messages {
		// Convert tool calls to repository format
		var toolCalls []repository.ToolCall
		for _, tc := range msg.ToolCalls {
			toolCalls = append(toolCalls, repository.ToolCall{
				ID:         tc.ID,
				Name:       tc.Name,
				Parameters: tc.Parameters,
			})
		}

		_, err := s.messageRepo.Create(conversationID, msg.Role, msg.Content, toolCalls, msg.ToolCallID)
		if err != nil {
			return fmt.Errorf("failed to save message: %w", err)
		}
	}

	return nil
}

// TrackTokens records token usage for the execution (Step 7 continued)
func (s *PersistStep) TrackTokens(ctx context.Context, executionID, userID, provider, model string, usage *llm.Usage) error {
	if usage == nil || s.tokenUsageRepo == nil {
		return nil
	}

	record := &workflow.TokenUsageRecord{
		ExecutionID:      executionID,
		UserID:           userID,
		Provider:         provider,
		Model:            model,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		CostUSD:          calculateCost(provider, model, usage),
	}

	err := s.tokenUsageRepo.Create(ctx, executionID, userID, provider, model, record)
	if err != nil {
		return fmt.Errorf("failed to track token usage: %w", err)
	}

	return nil
}

// UpdateExecutionIterations updates the iteration count for an execution
func (s *PersistStep) UpdateExecutionIterations(ctx context.Context, executionID string, iterations int) error {
	if s.executionRepo == nil {
		return nil
	}

	// Get current execution
	exec, err := s.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("failed to get execution: %w", err)
	}

	if exec == nil {
		return nil
	}

	// Update with new iteration count
	return s.executionRepo.Complete(ctx, executionID, exec.CommitSHA, iterations)
}

// calculateCost estimates the cost based on provider pricing
func calculateCost(provider, model string, usage *llm.Usage) float64 {
	if usage == nil {
		return 0
	}

	// Pricing per 1K tokens (approximate, as of 2024)
	var promptPrice, completionPrice float64

	switch provider {
	case "anthropic":
		switch {
		case contains(model, "claude-3-opus"):
			promptPrice = 0.015
			completionPrice = 0.075
		case contains(model, "claude-3-sonnet"):
			promptPrice = 0.003
			completionPrice = 0.015
		case contains(model, "claude-3-haiku"):
			promptPrice = 0.00025
			completionPrice = 0.00125
		default:
			promptPrice = 0.003
			completionPrice = 0.015
		}
	case "openai":
		switch {
		case contains(model, "gpt-4-turbo"), contains(model, "gpt-4o"):
			promptPrice = 0.01
			completionPrice = 0.03
		case contains(model, "gpt-4"):
			promptPrice = 0.03
			completionPrice = 0.06
		case contains(model, "gpt-3.5"):
			promptPrice = 0.0005
			completionPrice = 0.0015
		default:
			promptPrice = 0.01
			completionPrice = 0.03
		}
	case "google":
		switch {
		case contains(model, "gemini-pro"):
			promptPrice = 0.00025
			completionPrice = 0.0005
		case contains(model, "gemini-ultra"):
			promptPrice = 0.00125
			completionPrice = 0.00375
		default:
			promptPrice = 0.00025
			completionPrice = 0.0005
		}
	default:
		// Unknown provider, assume free (e.g., Ollama)
		return 0
	}

	promptCost := float64(usage.PromptTokens) / 1000 * promptPrice
	completionCost := float64(usage.CompletionTokens) / 1000 * completionPrice

	return promptCost + completionCost
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

// findSubstring finds substring index (simple implementation)
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
