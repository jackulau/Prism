package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/workflow"
)

// LLMStep handles LLM execution loop
type LLMStep struct {
	llmManager *llm.Manager
	emitter    *workflow.EventEmitter
}

// NewLLMStep creates a new LLM step handler
func NewLLMStep(llmManager *llm.Manager, emitter *workflow.EventEmitter) *LLMStep {
	return &LLMStep{
		llmManager: llmManager,
		emitter:    emitter,
	}
}

// RunLLMLoop executes the LLM agent loop with tool calling (Step 6)
func (s *LLMStep) RunLLMLoop(ctx context.Context, config workflow.LLMLoopConfig) (*workflow.LLMResult, error) {
	if config.Agent == nil {
		return nil, workflow.ErrInvalidAgentConfig
	}

	result := &workflow.LLMResult{
		Messages:     make([]llm.Message, 0),
		ToolCalls:    make([]workflow.ToolCallResult, 0),
		FilesChanged: make([]string, 0),
		Metadata:     make(map[string]interface{}),
	}

	// Copy initial messages
	messages := make([]llm.Message, len(config.Messages))
	copy(messages, config.Messages)

	// Add system prompt if not already present
	if config.Agent.SystemPrompt != "" && (len(messages) == 0 || messages[0].Role != "system") {
		systemMsg := llm.Message{
			Role:    "system",
			Content: config.Agent.SystemPrompt,
		}
		messages = append([]llm.Message{systemMsg}, messages...)
	}

	// Track total usage
	var totalUsage llm.Usage
	iteration := 0

	for {
		iteration++
		if iteration > config.MaxIterations {
			return nil, workflow.ErrMaxIterationsReached
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Emit progress
		if s.emitter != nil {
			s.emitter.EmitProgress(workflow.StepRunLLM, float64(iteration)/float64(config.MaxIterations),
				fmt.Sprintf("LLM iteration %d/%d", iteration, config.MaxIterations))
		}

		// Create chat request
		req := &llm.ChatRequest{
			Model:       config.Agent.Model,
			Messages:    messages,
			Tools:       config.Tools,
			Temperature: 0.7, // Default temperature
			Stream:      true,
		}

		// Execute LLM call
		stream, err := s.llmManager.Chat(ctx, config.Agent.Provider, req)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", workflow.ErrLLMExecutionFailed, err)
		}

		// Process streaming response
		var fullResponse string
		var toolCalls []llm.ToolCall
		var chunkUsage *llm.Usage

		for chunk := range stream {
			if chunk.Error != nil {
				return nil, fmt.Errorf("%w: %v", workflow.ErrLLMExecutionFailed, chunk.Error)
			}

			if chunk.Delta != "" {
				fullResponse += chunk.Delta
				if s.emitter != nil {
					s.emitter.EmitLLMChunk(chunk.Delta, nil)
				}
			}

			if len(chunk.ToolCalls) > 0 {
				toolCalls = append(toolCalls, chunk.ToolCalls...)
			}

			if chunk.Usage != nil {
				chunkUsage = chunk.Usage
			}
		}

		// Accumulate usage
		if chunkUsage != nil {
			totalUsage.PromptTokens += chunkUsage.PromptTokens
			totalUsage.CompletionTokens += chunkUsage.CompletionTokens
			totalUsage.TotalTokens += chunkUsage.TotalTokens
		}

		// Add assistant message to history
		assistantMsg := llm.Message{
			Role:      "assistant",
			Content:   fullResponse,
			ToolCalls: toolCalls,
		}
		messages = append(messages, assistantMsg)

		// If no tool calls, we're done
		if len(toolCalls) == 0 {
			result.Output = fullResponse
			break
		}

		// Execute tool calls
		for _, tc := range toolCalls {
			toolResult, err := s.executeToolCall(ctx, tc, config)
			if err != nil {
				// Record the error but continue
				toolResult = workflow.ToolCallResult{
					ToolCallID: tc.ID,
					Name:       tc.Name,
					Parameters: tc.Parameters,
					Error:      err.Error(),
				}
			}
			result.ToolCalls = append(result.ToolCalls, toolResult)

			// Track file changes
			if toolResult.Name == "write_file" || toolResult.Name == "edit_file" {
				if path, ok := tc.Parameters["path"].(string); ok {
					result.FilesChanged = append(result.FilesChanged, path)
				}
			}

			// Add tool result message
			toolMsg := llm.Message{
				Role:       "tool",
				Content:    toolResult.Output,
				ToolCallID: tc.ID,
			}
			if toolResult.Error != "" {
				toolMsg.Content = fmt.Sprintf("Error: %s", toolResult.Error)
			}
			messages = append(messages, toolMsg)
		}
	}

	result.Iterations = iteration
	result.TokenUsage = &totalUsage
	result.Messages = messages

	return result, nil
}

// executeToolCall executes a single tool call
func (s *LLMStep) executeToolCall(ctx context.Context, tc llm.ToolCall, config workflow.LLMLoopConfig) (workflow.ToolCallResult, error) {
	startTime := time.Now()

	result := workflow.ToolCallResult{
		ToolCallID: tc.ID,
		Name:       tc.Name,
		Parameters: tc.Parameters,
	}

	// Emit tool start event
	if s.emitter != nil {
		s.emitter.EmitToolExecution(tc.Name, "started", "", nil)
	}

	// Execute via tool executor if available
	if config.ToolExecutor != nil {
		output, err := config.ToolExecutor.Execute(ctx, tc, config.SandboxCtx)
		result.Duration = time.Since(startTime)

		if err != nil {
			result.Error = err.Error()
			if s.emitter != nil {
				s.emitter.EmitToolExecution(tc.Name, "failed", "", err)
			}
			return result, err
		}

		result.Output = output
		if s.emitter != nil {
			s.emitter.EmitToolExecution(tc.Name, "completed", output, nil)
		}
		return result, nil
	}

	// No tool executor available
	result.Duration = time.Since(startTime)
	result.Error = "no tool executor configured"
	return result, fmt.Errorf("no tool executor configured")
}

// BuildSystemPrompt builds the system prompt for the agent
func BuildSystemPrompt(agent *workflow.AgentData, sandboxCtx *workflow.SandboxContext) string {
	prompt := agent.SystemPrompt

	// Add context about the working directory
	if sandboxCtx != nil && sandboxCtx.WorkDir != "" {
		prompt += fmt.Sprintf("\n\nYou are working in directory: %s", sandboxCtx.WorkDir)
	}

	// Add context about the branch
	if sandboxCtx != nil && sandboxCtx.BranchName != "" {
		prompt += fmt.Sprintf("\nYou are working on branch: %s", sandboxCtx.BranchName)
	}

	// Add current task if available
	if agent.CurrentTask != "" {
		prompt += fmt.Sprintf("\n\nCurrent task: %s", agent.CurrentTask)
	}

	return prompt
}
