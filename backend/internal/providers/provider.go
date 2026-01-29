package providers

import (
	"context"
	"time"
)

// AgentProvider defines the interface for agent providers (Prism, Cursor, Jules, etc.)
// This is distinct from LLM providers - agent providers manage complete agent lifecycles
// including creation, messaging, and state management.
type AgentProvider interface {
	// Name returns the provider name (e.g., "prism", "cursor", "jules")
	Name() string

	// CreateAgent creates a new agent with the given request
	CreateAgent(ctx context.Context, req CreateAgentRequest) (*Agent, error)

	// GetAgent retrieves an agent by ID
	GetAgent(ctx context.Context, agentID string) (*Agent, error)

	// SendMessage sends a message to an agent and returns a streaming response
	SendMessage(ctx context.Context, agentID string, message string) (<-chan StreamChunk, error)

	// GetMessages retrieves the message history for an agent
	GetMessages(ctx context.Context, agentID string) ([]Message, error)

	// StopAgent stops a running agent
	StopAgent(ctx context.Context, agentID string) error

	// SupportsStreaming returns whether the provider supports streaming responses
	SupportsStreaming() bool
}

// CreateAgentRequest represents a request to create a new agent
type CreateAgentRequest struct {
	// Prompt is the initial prompt/task for the agent
	Prompt string `json:"prompt"`

	// Model is the optional model to use (provider-specific)
	Model string `json:"model,omitempty"`

	// WorkspaceID is the optional workspace context
	WorkspaceID string `json:"workspace_id,omitempty"`

	// UserID is the user making the request
	UserID string `json:"user_id,omitempty"`

	// Metadata contains additional provider-specific options
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Agent represents an agent instance
type Agent struct {
	// ID is the unique identifier for the agent
	ID string `json:"id"`

	// Provider is the provider that manages this agent
	Provider string `json:"provider"`

	// Status is the current agent status
	Status AgentStatus `json:"status"`

	// CreatedAt is when the agent was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the agent was last updated
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// Metadata contains additional provider-specific data
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AgentStatus represents the status of an agent
type AgentStatus string

const (
	AgentStatusPending   AgentStatus = "pending"
	AgentStatusRunning   AgentStatus = "running"
	AgentStatusCompleted AgentStatus = "completed"
	AgentStatusFailed    AgentStatus = "failed"
	AgentStatusStopped   AgentStatus = "stopped"
)

// Message represents a message in an agent conversation
type Message struct {
	// ID is the unique identifier for the message
	ID string `json:"id"`

	// Role is the message role (user, assistant, system, tool)
	Role string `json:"role"`

	// Content is the message content
	Content string `json:"content"`

	// CreatedAt is when the message was created
	CreatedAt time.Time `json:"created_at"`

	// ToolCalls contains any tool calls made in this message
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID is set if this is a tool response
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// ToolCall represents a tool call from an agent
type ToolCall struct {
	// ID is the tool call identifier
	ID string `json:"id"`

	// Name is the tool name
	Name string `json:"name"`

	// Parameters are the tool parameters
	Parameters map[string]interface{} `json:"parameters"`
}

// StreamChunk represents a chunk of streaming response from an agent
type StreamChunk struct {
	// Delta is the text delta for this chunk
	Delta string `json:"delta,omitempty"`

	// ToolCalls contains any tool calls in this chunk
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// FinishReason is set on the final chunk
	FinishReason string `json:"finish_reason,omitempty"`

	// Error contains any error that occurred
	Error error `json:"error,omitempty"`

	// MessageID is the ID of the message being streamed
	MessageID string `json:"message_id,omitempty"`
}

// ProviderError represents an error from a provider
type ProviderError struct {
	// Provider is the provider that returned the error
	Provider string `json:"provider"`

	// Code is the error code
	Code string `json:"code"`

	// Message is the error message
	Message string `json:"message"`

	// StatusCode is the HTTP status code (if applicable)
	StatusCode int `json:"status_code,omitempty"`

	// Retryable indicates if the request can be retried
	Retryable bool `json:"retryable"`
}

func (e *ProviderError) Error() string {
	return e.Message
}
