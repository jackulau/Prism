package prism

import (
	"testing"

	"github.com/jacklau/prism/internal/agent"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/providers"
)

func TestStreamAdapterConvertAgentEvent(t *testing.T) {
	adapter := NewStreamAdapter("anthropic", "claude-sonnet-4-5-20250929")

	tests := []struct {
		name          string
		event         *agent.AgentEvent
		expectedType  providers.StreamChunkType
		expectedDelta string
		expectNil     bool
	}{
		{
			name: "Started event",
			event: &agent.AgentEvent{
				Type: agent.AgentEventStarted,
			},
			expectedType: providers.StreamChunkTypeText,
		},
		{
			name: "Stream chunk event",
			event: &agent.AgentEvent{
				Type: agent.AgentEventStreamChunk,
				Data: map[string]interface{}{
					"delta": "Hello world",
				},
			},
			expectedType:  providers.StreamChunkTypeText,
			expectedDelta: "Hello world",
		},
		{
			name: "Tool call event",
			event: &agent.AgentEvent{
				Type: agent.AgentEventToolCall,
				Data: map[string]interface{}{
					"tool_call_id": "tc_123",
					"name":         "read_file",
					"parameters":   map[string]interface{}{"path": "/test.txt"},
				},
			},
			expectedType: providers.StreamChunkTypeToolCall,
		},
		{
			name: "Tool result event",
			event: &agent.AgentEvent{
				Type: agent.AgentEventToolResult,
				Data: map[string]interface{}{
					"tool_call_id": "tc_123",
					"name":         "read_file",
					"output":       "file contents",
				},
			},
			expectedType: providers.StreamChunkTypeToolResult,
		},
		{
			name: "Completed event",
			event: &agent.AgentEvent{
				Type: agent.AgentEventCompleted,
			},
			expectedType: providers.StreamChunkTypeDone,
		},
		{
			name: "Failed event",
			event: &agent.AgentEvent{
				Type: agent.AgentEventFailed,
				Data: map[string]interface{}{
					"error": "something went wrong",
				},
			},
			expectedType: providers.StreamChunkTypeError,
		},
		{
			name: "Cancelled event",
			event: &agent.AgentEvent{
				Type: agent.AgentEventCancelled,
			},
			expectedType: providers.StreamChunkTypeError,
		},
		{
			name:      "Nil event",
			event:     nil,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := adapter.ConvertAgentEvent(tt.event)

			if tt.expectNil {
				if chunk != nil {
					t.Errorf("expected nil chunk, got %+v", chunk)
				}
				return
			}

			if chunk == nil {
				t.Fatal("expected non-nil chunk")
			}

			if chunk.Type != tt.expectedType {
				t.Errorf("expected type %s, got %s", tt.expectedType, chunk.Type)
			}

			if tt.expectedDelta != "" && chunk.Delta != tt.expectedDelta {
				t.Errorf("expected delta '%s', got '%s'", tt.expectedDelta, chunk.Delta)
			}
		})
	}
}

func TestStreamAdapterConvertLLMChunk(t *testing.T) {
	adapter := NewStreamAdapter("anthropic", "claude-sonnet-4-5-20250929")

	tests := []struct {
		name          string
		chunk         llm.StreamChunk
		expectedType  providers.StreamChunkType
		expectedDelta string
		expectNil     bool
	}{
		{
			name: "Text delta",
			chunk: llm.StreamChunk{
				Delta: "Hello ",
			},
			expectedType:  providers.StreamChunkTypeText,
			expectedDelta: "Hello ",
		},
		{
			name: "Tool call",
			chunk: llm.StreamChunk{
				ToolCalls: []llm.ToolCall{
					{
						ID:         "tc_1",
						Name:       "bash",
						Parameters: map[string]interface{}{"command": "ls"},
					},
				},
			},
			expectedType: providers.StreamChunkTypeToolCall,
		},
		{
			name: "Finish with usage",
			chunk: llm.StreamChunk{
				FinishReason: "stop",
				Usage: &llm.Usage{
					PromptTokens:     100,
					CompletionTokens: 50,
					TotalTokens:      150,
				},
			},
			expectedType: providers.StreamChunkTypeDone,
		},
		{
			name: "Usage only",
			chunk: llm.StreamChunk{
				Usage: &llm.Usage{
					PromptTokens:     100,
					CompletionTokens: 50,
					TotalTokens:      150,
				},
			},
			expectedType: providers.StreamChunkTypeUsage,
		},
		{
			name: "Error",
			chunk: llm.StreamChunk{
				Error: &StreamError{Message: "API error"},
			},
			expectedType: providers.StreamChunkTypeError,
		},
		{
			name:      "Empty chunk",
			chunk:     llm.StreamChunk{},
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := adapter.ConvertLLMChunk(tt.chunk)

			if tt.expectNil {
				if chunk != nil {
					t.Errorf("expected nil chunk, got %+v", chunk)
				}
				return
			}

			if chunk == nil {
				t.Fatal("expected non-nil chunk")
			}

			if chunk.Type != tt.expectedType {
				t.Errorf("expected type %s, got %s", tt.expectedType, chunk.Type)
			}

			if tt.expectedDelta != "" && chunk.Delta != tt.expectedDelta {
				t.Errorf("expected delta '%s', got '%s'", tt.expectedDelta, chunk.Delta)
			}
		})
	}
}

func TestStreamAdapterConvertAgentResult(t *testing.T) {
	adapter := NewStreamAdapter("anthropic", "claude-sonnet-4-5-20250929")

	tests := []struct {
		name           string
		result         *agent.AgentResult
		expectedType   providers.StreamChunkType
		expectUsage    bool
		expectCost     bool
		expectError    bool
	}{
		{
			name:         "Nil result",
			result:       nil,
			expectedType: providers.StreamChunkTypeDone,
		},
		{
			name: "Successful result with usage",
			result: &agent.AgentResult{
				Success: true,
				Usage: &llm.Usage{
					PromptTokens:     1000,
					CompletionTokens: 500,
					TotalTokens:      1500,
				},
			},
			expectedType: providers.StreamChunkTypeDone,
			expectUsage:  true,
			expectCost:   true,
		},
		{
			name: "Failed result",
			result: &agent.AgentResult{
				Success: false,
				Error:   "execution failed",
			},
			expectedType: providers.StreamChunkTypeError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := adapter.ConvertAgentResult(tt.result)

			if chunk == nil {
				t.Fatal("expected non-nil chunk")
			}

			if chunk.Type != tt.expectedType {
				t.Errorf("expected type %s, got %s", tt.expectedType, chunk.Type)
			}

			if tt.expectUsage && chunk.Usage == nil {
				t.Error("expected usage, got nil")
			}

			if tt.expectCost && chunk.Cost == nil {
				t.Error("expected cost, got nil")
			}

			if tt.expectError && chunk.Error == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestAggregateUsage(t *testing.T) {
	usage1 := &providers.Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	usage2 := &providers.Usage{
		PromptTokens:     200,
		CompletionTokens: 100,
		TotalTokens:      300,
	}

	result := AggregateUsage(usage1, usage2, nil)

	if result.PromptTokens != 300 {
		t.Errorf("expected 300 prompt tokens, got %d", result.PromptTokens)
	}

	if result.CompletionTokens != 150 {
		t.Errorf("expected 150 completion tokens, got %d", result.CompletionTokens)
	}

	if result.TotalTokens != 450 {
		t.Errorf("expected 450 total tokens, got %d", result.TotalTokens)
	}
}

func TestAggregateCost(t *testing.T) {
	cost1 := &providers.Cost{
		InputCost:  0.10,
		OutputCost: 0.20,
		TotalCost:  0.30,
		Currency:   "USD",
	}

	cost2 := &providers.Cost{
		InputCost:  0.15,
		OutputCost: 0.25,
		TotalCost:  0.40,
		Currency:   "USD",
	}

	result := AggregateCost(cost1, cost2, nil)

	if result.InputCost != 0.25 {
		t.Errorf("expected 0.25 input cost, got %f", result.InputCost)
	}

	if result.OutputCost != 0.45 {
		t.Errorf("expected 0.45 output cost, got %f", result.OutputCost)
	}

	if result.TotalCost != 0.70 {
		t.Errorf("expected 0.70 total cost, got %f", result.TotalCost)
	}

	if result.Currency != "USD" {
		t.Errorf("expected USD currency, got %s", result.Currency)
	}
}

func TestStreamError(t *testing.T) {
	err := &StreamError{Message: "test error"}

	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got '%s'", err.Error())
	}
}
