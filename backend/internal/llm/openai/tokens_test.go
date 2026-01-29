package openai

import (
	"testing"

	"github.com/jacklau/prism/internal/llm"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "short text",
			text:     "Hello",
			expected: 1, // 5 chars / 4 = 1
		},
		{
			name:     "longer text",
			text:     "This is a longer piece of text that should have more tokens",
			expected: 14, // 59 chars / 4 = 14
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTokens(tt.text)
			if got != tt.expected {
				t.Errorf("EstimateTokens(%q) = %d, want %d", tt.text, got, tt.expected)
			}
		})
	}
}

func TestEstimateMessagesTokens(t *testing.T) {
	tests := []struct {
		name     string
		messages []llm.Message
		minTokens int // Use min since estimation is approximate
	}{
		{
			name:     "empty messages",
			messages: []llm.Message{},
			minTokens: 3, // Base overhead
		},
		{
			name: "single message",
			messages: []llm.Message{
				{Role: "user", Content: "Hello"},
			},
			minTokens: 5, // Base + message overhead + content
		},
		{
			name: "multiple messages",
			messages: []llm.Message{
				{Role: "system", Content: "You are a helpful assistant"},
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
			},
			minTokens: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateMessagesTokens(tt.messages)
			if got < tt.minTokens {
				t.Errorf("EstimateMessagesTokens() = %d, want at least %d", got, tt.minTokens)
			}
		})
	}
}

func TestTokenCounter_CountTokens(t *testing.T) {
	tc, err := NewTokenCounter("gpt-4")
	if err != nil {
		t.Skipf("tiktoken not available: %v", err)
	}

	tests := []struct {
		name string
		text string
		min  int
		max  int
	}{
		{
			name: "empty string",
			text: "",
			min:  0,
			max:  0,
		},
		{
			name: "hello world",
			text: "Hello, world!",
			min:  1,
			max:  10,
		},
		{
			name: "longer text",
			text: "The quick brown fox jumps over the lazy dog.",
			min:  5,
			max:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tc.CountTokens(tt.text)
			if got < tt.min || got > tt.max {
				t.Errorf("CountTokens(%q) = %d, want between %d and %d", tt.text, got, tt.min, tt.max)
			}
		})
	}
}

func TestTokenCounter_CountMessage(t *testing.T) {
	tc, err := NewTokenCounter("gpt-4")
	if err != nil {
		t.Skipf("tiktoken not available: %v", err)
	}

	msg := llm.Message{
		Role:    "user",
		Content: "Hello, how are you?",
	}

	tokens := tc.CountMessage(msg)
	if tokens < 5 {
		t.Errorf("CountMessage() = %d, want at least 5 tokens", tokens)
	}
}

func TestTokenCounter_CountMessages(t *testing.T) {
	tc, err := NewTokenCounter("gpt-4")
	if err != nil {
		t.Skipf("tiktoken not available: %v", err)
	}

	messages := []llm.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "What is 2+2?"},
		{Role: "assistant", Content: "2+2 equals 4."},
	}

	tokens := tc.CountMessages(messages)
	if tokens < 15 {
		t.Errorf("CountMessages() = %d, want at least 15 tokens", tokens)
	}
}

func TestTokenCounter_CountToolDefinition(t *testing.T) {
	tc, err := NewTokenCounter("gpt-4")
	if err != nil {
		t.Skipf("tiktoken not available: %v", err)
	}

	tool := llm.ToolDefinition{
		Name:        "get_weather",
		Description: "Get the current weather for a location",
		Parameters: llm.JSONSchema{
			Type: "object",
			Properties: map[string]llm.JSONProperty{
				"location": {
					Type:        "string",
					Description: "The city and state",
				},
			},
			Required: []string{"location"},
		},
	}

	tokens := tc.CountToolDefinition(tool)
	if tokens < 10 {
		t.Errorf("CountToolDefinition() = %d, want at least 10 tokens", tokens)
	}
}

func TestTokenCounter_CountRequest(t *testing.T) {
	tc, err := NewTokenCounter("gpt-4")
	if err != nil {
		t.Skipf("tiktoken not available: %v", err)
	}

	req := &llm.ChatRequest{
		Model: "gpt-4",
		Messages: []llm.Message{
			{Role: "user", Content: "What's the weather like?"},
		},
		Tools: []llm.ToolDefinition{
			{
				Name:        "get_weather",
				Description: "Get weather info",
				Parameters: llm.JSONSchema{
					Type: "object",
				},
			},
		},
	}

	tokens := tc.CountRequest(req)
	if tokens < 10 {
		t.Errorf("CountRequest() = %d, want at least 10 tokens", tokens)
	}
}

func TestTokenCounter_GetContextUsage(t *testing.T) {
	tc, err := NewTokenCounter("gpt-4o")
	if err != nil {
		t.Skipf("tiktoken not available: %v", err)
	}

	// GPT-4o has 128000 context window
	usage := tc.GetContextUsage(64000)
	if usage != 50.0 {
		t.Errorf("GetContextUsage(64000) = %f, want 50.0", usage)
	}

	usage = tc.GetContextUsage(128000)
	if usage != 100.0 {
		t.Errorf("GetContextUsage(128000) = %f, want 100.0", usage)
	}
}

func TestTokenCounter_RemainingTokens(t *testing.T) {
	tc, err := NewTokenCounter("gpt-4o")
	if err != nil {
		t.Skipf("tiktoken not available: %v", err)
	}

	// GPT-4o has 16384 max output
	remaining := tc.RemainingTokens(100000)
	if remaining > 16384 {
		t.Errorf("RemainingTokens(100000) = %d, should be capped at max output", remaining)
	}

	// When input is small, remaining should be max output
	remaining = tc.RemainingTokens(1000)
	if remaining != 16384 {
		t.Errorf("RemainingTokens(1000) = %d, want 16384", remaining)
	}
}

func TestEncodingCache(t *testing.T) {
	// Test that encoding is cached and reused
	tc1, err := NewTokenCounter("gpt-4")
	if err != nil {
		t.Skipf("tiktoken not available: %v", err)
	}

	tc2, err := NewTokenCounter("gpt-4")
	if err != nil {
		t.Fatalf("Failed to create second counter: %v", err)
	}

	// Both should work correctly
	tokens1 := tc1.CountTokens("test")
	tokens2 := tc2.CountTokens("test")

	if tokens1 != tokens2 {
		t.Errorf("Different token counts from cached encodings: %d vs %d", tokens1, tokens2)
	}
}
