package cloudprovider

import (
	"time"
)

// AgentStatus represents the status of a cloud agent
type AgentStatus string

const (
	AgentStatusActive     AgentStatus = "active"
	AgentStatusIdle       AgentStatus = "idle"
	AgentStatusProcessing AgentStatus = "processing"
	AgentStatusTerminated AgentStatus = "terminated"
)

// Agent represents a cloud-hosted AI agent
type Agent struct {
	// ID is the unique identifier for this agent
	ID string `json:"id"`

	// ProviderID is the ID assigned by the cloud provider
	ProviderID string `json:"provider_id"`

	// Name is the human-readable name of the agent
	Name string `json:"name"`

	// Model is the model identifier used by this agent
	Model string `json:"model"`

	// Status is the current status of the agent
	Status AgentStatus `json:"status"`

	// SystemPrompt is the system instruction for the agent
	SystemPrompt string `json:"system_prompt,omitempty"`

	// CreatedAt is when the agent was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the agent was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Metadata contains provider-specific metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CreateAgentParams contains parameters for creating a new agent
type CreateAgentParams struct {
	// Name is the human-readable name for the agent
	Name string `json:"name"`

	// Model is the model identifier to use
	Model string `json:"model"`

	// SystemPrompt is the system instruction for the agent
	SystemPrompt string `json:"system_prompt,omitempty"`

	// Tools is a list of tool definitions available to the agent
	Tools []ToolDefinition `json:"tools,omitempty"`

	// Metadata contains provider-specific options
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToolDefinition defines a tool available to the agent
type ToolDefinition struct {
	// Name is the tool identifier
	Name string `json:"name"`

	// Description explains what the tool does
	Description string `json:"description"`

	// Parameters defines the JSON schema for tool parameters
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ProviderMessage represents a message in an agent conversation
type ProviderMessage struct {
	// ID is the unique identifier for this message
	ID string `json:"id"`

	// Role is the message role (user, assistant, tool)
	Role string `json:"role"`

	// Content is the text content of the message
	Content string `json:"content"`

	// ToolCalls contains any tool calls made in this message
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID is the ID of the tool call this message responds to
	ToolCallID string `json:"tool_call_id,omitempty"`

	// Timestamp is when the message was created
	Timestamp time.Time `json:"timestamp"`
}

// ToolCall represents a tool invocation by the agent
type ToolCall struct {
	// ID is the unique identifier for this tool call
	ID string `json:"id"`

	// Name is the name of the tool being called
	Name string `json:"name"`

	// Parameters contains the tool call parameters
	Parameters map[string]interface{} `json:"parameters"`
}

// ImageData represents image data to send with a message
type ImageData struct {
	// URL is the URL of the image (mutually exclusive with Base64)
	URL string `json:"url,omitempty"`

	// Base64 is the base64-encoded image data
	Base64 string `json:"base64,omitempty"`

	// MimeType is the MIME type of the image (e.g., "image/png")
	MimeType string `json:"mime_type,omitempty"`
}

// MessageChunk represents a chunk of a streaming message response
type MessageChunk struct {
	// Delta is the text content delta
	Delta string `json:"delta,omitempty"`

	// ToolCalls contains any tool calls in this chunk
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// FinishReason indicates why streaming stopped (only on final chunk)
	FinishReason string `json:"finish_reason,omitempty"`

	// Error contains any error that occurred during streaming
	Error error `json:"error,omitempty"`

	// MessageID is the ID of the message being streamed
	MessageID string `json:"message_id,omitempty"`
}

// ProviderConfig holds configuration for a cloud provider
type ProviderConfig struct {
	// APIKey is the API key for authentication
	APIKey string `json:"api_key"`

	// BaseURL is an optional custom base URL
	BaseURL string `json:"base_url,omitempty"`

	// OrgID is an optional organization ID
	OrgID string `json:"org_id,omitempty"`
}
