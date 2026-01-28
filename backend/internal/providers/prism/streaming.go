package prism

import (
	"github.com/jacklau/prism/internal/agent"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/providers"
)

// StreamAdapter converts agent events to provider stream chunks
type StreamAdapter struct {
	provider string
	model    string
}

// NewStreamAdapter creates a new stream adapter
func NewStreamAdapter(provider, model string) *StreamAdapter {
	return &StreamAdapter{
		provider: provider,
		model:    model,
	}
}

// ConvertAgentEvent converts an agent event to a provider stream chunk
func (s *StreamAdapter) ConvertAgentEvent(event *agent.AgentEvent) *providers.StreamChunk {
	if event == nil {
		return nil
	}

	switch event.Type {
	case agent.AgentEventStarted:
		return &providers.StreamChunk{
			Type: providers.StreamChunkTypeText,
		}

	case agent.AgentEventStreamChunk:
		delta, _ := event.Data["delta"].(string)
		return &providers.StreamChunk{
			Type:  providers.StreamChunkTypeText,
			Delta: delta,
		}

	case agent.AgentEventThinking:
		thinking, _ := event.Data["thinking"].(string)
		return &providers.StreamChunk{
			Type:  providers.StreamChunkTypeText,
			Delta: thinking,
		}

	case agent.AgentEventToolCall:
		toolCallID, _ := event.Data["tool_call_id"].(string)
		name, _ := event.Data["name"].(string)
		params, _ := event.Data["parameters"].(map[string]interface{})

		return &providers.StreamChunk{
			Type: providers.StreamChunkTypeToolCall,
			ToolCall: &providers.ToolCall{
				ID:         toolCallID,
				Name:       name,
				Parameters: params,
			},
		}

	case agent.AgentEventToolResult:
		toolCallID, _ := event.Data["tool_call_id"].(string)
		name, _ := event.Data["name"].(string)
		output, _ := event.Data["output"].(string)
		errStr, _ := event.Data["error"].(string)

		return &providers.StreamChunk{
			Type: providers.StreamChunkTypeToolResult,
			ToolResult: &providers.ToolResult{
				ToolCallID: toolCallID,
				Name:       name,
				Output:     output,
				Error:      errStr,
			},
		}

	case agent.AgentEventCompleted:
		return &providers.StreamChunk{
			Type: providers.StreamChunkTypeDone,
			Done: true,
		}

	case agent.AgentEventFailed:
		errMsg, _ := event.Data["error"].(string)
		return &providers.StreamChunk{
			Type:  providers.StreamChunkTypeError,
			Error: &StreamError{Message: errMsg},
		}

	case agent.AgentEventCancelled:
		return &providers.StreamChunk{
			Type:  providers.StreamChunkTypeError,
			Error: &StreamError{Message: "cancelled"},
		}

	default:
		return nil
	}
}

// ConvertLLMChunk converts an LLM stream chunk to a provider stream chunk
func (s *StreamAdapter) ConvertLLMChunk(chunk llm.StreamChunk) *providers.StreamChunk {
	// Handle error
	if chunk.Error != nil {
		return &providers.StreamChunk{
			Type:  providers.StreamChunkTypeError,
			Error: chunk.Error,
		}
	}

	// Handle text delta
	if chunk.Delta != "" {
		return &providers.StreamChunk{
			Type:  providers.StreamChunkTypeText,
			Delta: chunk.Delta,
		}
	}

	// Handle tool calls
	if len(chunk.ToolCalls) > 0 {
		tc := chunk.ToolCalls[0] // Process one at a time
		return &providers.StreamChunk{
			Type: providers.StreamChunkTypeToolCall,
			ToolCall: &providers.ToolCall{
				ID:         tc.ID,
				Name:       tc.Name,
				Parameters: tc.Parameters,
			},
		}
	}

	// Handle finish
	if chunk.FinishReason != "" {
		result := &providers.StreamChunk{
			Type: providers.StreamChunkTypeDone,
			Done: true,
		}

		// Add usage if available
		if chunk.Usage != nil {
			result.Usage = &providers.Usage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
			}

			// Calculate cost
			result.Cost = CalculateCost(s.provider, s.model, chunk.Usage)
		}

		return result
	}

	// Handle usage without finish
	if chunk.Usage != nil {
		return &providers.StreamChunk{
			Type: providers.StreamChunkTypeUsage,
			Usage: &providers.Usage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
			},
			Cost: CalculateCost(s.provider, s.model, chunk.Usage),
		}
	}

	return nil
}

// ConvertAgentResult converts an agent result to a final stream chunk with usage
func (s *StreamAdapter) ConvertAgentResult(result *agent.AgentResult) *providers.StreamChunk {
	if result == nil {
		return &providers.StreamChunk{
			Type: providers.StreamChunkTypeDone,
			Done: true,
		}
	}

	chunk := &providers.StreamChunk{
		Type: providers.StreamChunkTypeDone,
		Done: true,
	}

	if result.Usage != nil {
		chunk.Usage = &providers.Usage{
			PromptTokens:     result.Usage.PromptTokens,
			CompletionTokens: result.Usage.CompletionTokens,
			TotalTokens:      result.Usage.TotalTokens,
		}
		chunk.Cost = CalculateCost(s.provider, s.model, result.Usage)
	}

	if !result.Success && result.Error != "" {
		chunk.Type = providers.StreamChunkTypeError
		chunk.Error = &StreamError{Message: result.Error}
	}

	return chunk
}

// StreamError implements the error interface for stream errors
type StreamError struct {
	Message string
}

func (e *StreamError) Error() string {
	return e.Message
}

// AggregateUsage aggregates multiple usage records into one
func AggregateUsage(usages ...*providers.Usage) *providers.Usage {
	result := &providers.Usage{}
	for _, u := range usages {
		if u != nil {
			result.PromptTokens += u.PromptTokens
			result.CompletionTokens += u.CompletionTokens
			result.TotalTokens += u.TotalTokens
		}
	}
	return result
}

// AggregateCost aggregates multiple cost records into one
func AggregateCost(costs ...*providers.Cost) *providers.Cost {
	result := &providers.Cost{
		Currency: "USD",
	}
	for _, c := range costs {
		if c != nil {
			result.InputCost += c.InputCost
			result.OutputCost += c.OutputCost
			result.TotalCost += c.TotalCost
		}
	}
	return result
}
