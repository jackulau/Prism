package llm

import (
	"context"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// Name returns the provider name (e.g., "openai", "anthropic")
	Name() string

	// Models returns the list of available models
	Models() []Model

	// Chat sends a chat request and returns a streaming response
	Chat(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error)

	// SupportsTools returns whether the provider supports tool calling
	SupportsTools() bool

	// SupportsVision returns whether the provider supports vision/images
	SupportsVision() bool

	// ValidateKey validates an API key
	ValidateKey(ctx context.Context, key string) error
}

// Model represents an available model
type Model struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	ContextWindow int      `json:"context_window"`
	SupportsTools bool     `json:"supports_tools"`
	SupportsVision bool    `json:"supports_vision"`
	Capabilities  []string `json:"capabilities,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role       string      `json:"role"` // "system", "user", "assistant", "tool"
	Content    string      `json:"content"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
	Images     []ImageData `json:"images,omitempty"`
}

// ImageData represents an image in a message
type ImageData struct {
	URL      string `json:"url,omitempty"`
	Base64   string `json:"base64,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

// ToolCall represents a tool call from the LLM
type ToolCall struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ToolDefinition defines a tool that can be called by the LLM
type ToolDefinition struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  JSONSchema `json:"parameters"`
}

// JSONSchema represents a JSON Schema for tool parameters
type JSONSchema struct {
	Type       string                   `json:"type"`
	Properties map[string]JSONProperty  `json:"properties,omitempty"`
	Required   []string                 `json:"required,omitempty"`
}

// JSONProperty represents a property in a JSON Schema
type JSONProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Default     any      `json:"default,omitempty"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string           `json:"model"`
	Messages    []Message        `json:"messages"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Stream      bool             `json:"stream"`
}

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	// Text delta
	Delta string `json:"delta,omitempty"`

	// Tool calls (if any)
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// Finish reason (only on final chunk)
	FinishReason string `json:"finish_reason,omitempty"`

	// Error (if any)
	Error error `json:"error,omitempty"`

	// Usage statistics (only on final chunk)
	Usage *Usage `json:"usage,omitempty"`
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ProviderConfig holds configuration for a provider
type ProviderConfig struct {
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url,omitempty"`
	OrgID    string `json:"org_id,omitempty"`
}
