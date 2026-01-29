package providers

import (
	"context"
	"time"

	"github.com/jacklau/prism/internal/llm"
)

// AgentProvider defines the interface for agent providers (Prism, Cursor, Jules, etc.)
// This is distinct from LLM providers - agent providers execute autonomous agents
// with full context, tool use, and state management capabilities.
type AgentProvider interface {
	// Name returns the unique identifier for this provider
	Name() string

	// CreateAgent creates a new agent with the given configuration
	// Returns the agent with a unique ID that can be used for subsequent operations
	CreateAgent(ctx context.Context, req CreateAgentRequest) (*Agent, error)

	// GetAgent retrieves an agent by its ID
	GetAgent(ctx context.Context, agentID string) (*Agent, error)

	// SendMessage sends a message to an agent and returns a streaming response
	// The channel emits StreamChunk events as the agent processes the message
	SendMessage(ctx context.Context, agentID string, message string) (<-chan StreamChunk, error)

	// GetMessages retrieves the message history for an agent
	GetMessages(ctx context.Context, agentID string) ([]Message, error)

	// StopAgent stops a running agent
	StopAgent(ctx context.Context, agentID string) error

	// SupportsStreaming returns whether the provider supports streaming responses
	SupportsStreaming() bool

	// Capabilities returns the provider's capabilities
	Capabilities() ProviderCapabilities
}

// CreateAgentRequest contains parameters for creating a new agent
type CreateAgentRequest struct {
	// UserID is the ID of the user creating the agent
	UserID string `json:"user_id"`

	// Name is an optional name for the agent
	Name string `json:"name,omitempty"`

	// Description describes the agent's purpose
	Description string `json:"description,omitempty"`

	// Provider is the LLM provider to use (e.g., "anthropic", "openai")
	Provider string `json:"provider"`

	// Model is the specific model to use
	Model string `json:"model"`

	// SystemPrompt is the system prompt for the agent
	SystemPrompt string `json:"system_prompt,omitempty"`

	// Tools is the list of tools available to the agent
	Tools []llm.ToolDefinition `json:"tools,omitempty"`

	// Temperature controls randomness (0-1)
	Temperature float64 `json:"temperature,omitempty"`

	// MaxTokens limits the response length
	MaxTokens int `json:"max_tokens,omitempty"`

	// Metadata contains additional provider-specific configuration
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Agent represents an agent instance managed by a provider
type Agent struct {
	// ID is the unique identifier for this agent
	ID string `json:"id"`

	// ProviderName identifies which provider manages this agent
	ProviderName string `json:"provider_name"`

	// UserID is the ID of the user who owns this agent
	UserID string `json:"user_id"`

	// Name is the optional name of the agent
	Name string `json:"name,omitempty"`

	// Description describes what the agent does
	Description string `json:"description,omitempty"`

	// Status is the current status of the agent
	Status AgentStatus `json:"status"`

	// LLMProvider is the underlying LLM provider (e.g., "anthropic")
	LLMProvider string `json:"llm_provider"`

	// Model is the model being used
	Model string `json:"model"`

	// CreatedAt is when the agent was created
	CreatedAt time.Time `json:"created_at"`

	// StartedAt is when the agent started processing (if applicable)
	StartedAt *time.Time `json:"started_at,omitempty"`

	// CompletedAt is when the agent completed (if applicable)
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Usage contains token usage statistics
	Usage *Usage `json:"usage,omitempty"`

	// Cost contains cost information
	Cost *Cost `json:"cost,omitempty"`

	// Error contains error information if the agent failed
	Error string `json:"error,omitempty"`

	// Metadata contains additional provider-specific data
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AgentStatus represents the current status of an agent
type AgentStatus string

const (
	AgentStatusIdle      AgentStatus = "idle"
	AgentStatusRunning   AgentStatus = "running"
	AgentStatusCompleted AgentStatus = "completed"
	AgentStatusFailed    AgentStatus = "failed"
	AgentStatusCancelled AgentStatus = "cancelled"
)

// Message represents a message in an agent's conversation
type Message struct {
	// ID is the unique identifier for this message
	ID string `json:"id"`

	// Role is the message role (user, assistant, system, tool)
	Role string `json:"role"`

	// Content is the text content of the message
	Content string `json:"content"`

	// ToolCalls contains any tool calls made by the assistant
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID is the ID of the tool call this message responds to
	ToolCallID string `json:"tool_call_id,omitempty"`

	// Timestamp is when the message was created
	Timestamp time.Time `json:"timestamp"`

	// Usage contains token usage for this message
	Usage *Usage `json:"usage,omitempty"`
}

// ToolCall represents a tool call made by the agent
type ToolCall struct {
	// ID is the unique identifier for this tool call
	ID string `json:"id"`

	// Name is the name of the tool being called
	Name string `json:"name"`

	// Parameters are the parameters passed to the tool
	Parameters map[string]interface{} `json:"parameters"`
}

// StreamChunk represents a chunk of streaming response from an agent
type StreamChunk struct {
	// Type indicates the type of chunk
	Type StreamChunkType `json:"type"`

	// Delta contains text content delta
	Delta string `json:"delta,omitempty"`

	// ToolCall contains tool call information
	ToolCall *ToolCall `json:"tool_call,omitempty"`

	// ToolResult contains tool execution result
	ToolResult *ToolResult `json:"tool_result,omitempty"`

	// Usage contains token usage (typically on final chunk)
	Usage *Usage `json:"usage,omitempty"`

	// Cost contains cost information (typically on final chunk)
	Cost *Cost `json:"cost,omitempty"`

	// Error contains error information
	Error error `json:"error,omitempty"`

	// Done indicates this is the final chunk
	Done bool `json:"done,omitempty"`
}

// StreamChunkType represents the type of stream chunk
type StreamChunkType string

const (
	StreamChunkTypeText       StreamChunkType = "text"
	StreamChunkTypeToolCall   StreamChunkType = "tool_call"
	StreamChunkTypeToolResult StreamChunkType = "tool_result"
	StreamChunkTypeUsage      StreamChunkType = "usage"
	StreamChunkTypeError      StreamChunkType = "error"
	StreamChunkTypeDone       StreamChunkType = "done"
)

// ToolResult represents the result of a tool execution
type ToolResult struct {
	// ToolCallID is the ID of the tool call this responds to
	ToolCallID string `json:"tool_call_id"`

	// Name is the name of the tool that was executed
	Name string `json:"name"`

	// Output is the output from the tool
	Output string `json:"output"`

	// Error contains any error from tool execution
	Error string `json:"error,omitempty"`

	// Duration is how long the tool took to execute
	Duration time.Duration `json:"duration,omitempty"`
}

// Usage contains token usage statistics
type Usage struct {
	// PromptTokens is the number of tokens in the prompt
	PromptTokens int `json:"prompt_tokens"`

	// CompletionTokens is the number of tokens in the completion
	CompletionTokens int `json:"completion_tokens"`

	// TotalTokens is the total number of tokens used
	TotalTokens int `json:"total_tokens"`
}

// Cost contains cost information for an agent execution
type Cost struct {
	// InputCost is the cost of input tokens
	InputCost float64 `json:"input_cost"`

	// OutputCost is the cost of output tokens
	OutputCost float64 `json:"output_cost"`

	// TotalCost is the total cost
	TotalCost float64 `json:"total_cost"`

	// Currency is the currency (typically "USD")
	Currency string `json:"currency"`
}

// ProviderCapabilities describes what a provider can do
type ProviderCapabilities struct {
	// Streaming indicates support for streaming responses
	Streaming bool `json:"streaming"`

	// Tools indicates support for tool/function calling
	Tools bool `json:"tools"`

	// Vision indicates support for image inputs
	Vision bool `json:"vision"`

	// MultiAgent indicates support for multi-agent orchestration
	MultiAgent bool `json:"multi_agent"`

	// Sandbox indicates support for sandboxed code execution
	Sandbox bool `json:"sandbox"`

	// CostTracking indicates support for cost tracking
	CostTracking bool `json:"cost_tracking"`

	// Persistence indicates support for conversation persistence
	Persistence bool `json:"persistence"`
}
