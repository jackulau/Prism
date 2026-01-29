package cloudprovider

import (
	"time"
)

// AgentStatus represents the current status of a cloud agent
type AgentStatus string

const (
	// AgentStatusActive indicates the agent is actively processing
	AgentStatusActive AgentStatus = "active"
	// AgentStatusIdle indicates the agent is idle and waiting for input
	AgentStatusIdle AgentStatus = "idle"
	// AgentStatusProcessing indicates the agent is processing a request
	AgentStatusProcessing AgentStatus = "processing"
	// AgentStatusTerminated indicates the agent has been terminated
	AgentStatusTerminated AgentStatus = "terminated"
	// AgentStatusError indicates the agent encountered an error
	AgentStatusError AgentStatus = "error"
)

// Agent represents a cloud-hosted AI agent
type Agent struct {
	// ID is the local identifier for this agent
	ID string `json:"id"`
	// ProviderID is the provider's identifier for this agent
	ProviderID string `json:"provider_id"`
	// Name is the display name for this agent
	Name string `json:"name"`
	// Provider is the name of the cloud provider (e.g., "claude-cloud")
	Provider string `json:"provider"`
	// Model is the model being used by this agent
	Model string `json:"model,omitempty"`
	// SystemPrompt is the system prompt/instructions for the agent
	SystemPrompt string `json:"system_prompt,omitempty"`
	// Status is the current status of the agent
	Status AgentStatus `json:"status"`
	// CreatedAt is when the agent was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the agent was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// Metadata holds provider-specific data
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CreateAgentParams contains parameters for creating a new agent
type CreateAgentParams struct {
	// Name is the display name for the agent
	Name string `json:"name"`
	// SystemPrompt is the system prompt/instructions for the agent
	SystemPrompt string `json:"system_prompt,omitempty"`
	// Model is the model to use (provider-specific)
	Model string `json:"model,omitempty"`
	// Tools is the list of tools the agent can use
	Tools []ToolDefinition `json:"tools,omitempty"`
	// MaxTokens is the maximum tokens for responses
	MaxTokens int `json:"max_tokens,omitempty"`
	// Metadata holds provider-specific configuration
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToolDefinition defines a tool that can be used by the agent
type ToolDefinition struct {
	// Name is the tool identifier
	Name string `json:"name"`
	// Description describes what the tool does
	Description string `json:"description"`
	// Parameters defines the tool's input schema (JSON Schema)
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ProviderMessage represents a message in an agent's conversation
type ProviderMessage struct {
	// ID is the provider's identifier for this message
	ID string `json:"id"`
	// Role is the message role ("user", "assistant", "system", "tool")
	Role string `json:"role"`
	// Content is the text content of the message
	Content string `json:"content"`
	// Timestamp is when the message was created
	Timestamp time.Time `json:"timestamp"`
	// ToolCalls contains any tool calls made in this message
	ToolCalls []ToolUseRequest `json:"tool_calls,omitempty"`
	// ToolCallID is set when this message is a tool result
	ToolCallID string `json:"tool_call_id,omitempty"`
	// Images contains any images in this message
	Images []ImageData `json:"images,omitempty"`
	// Metadata holds provider-specific data
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToolUseRequest represents a tool call made by the agent
type ToolUseRequest struct {
	// ID is the unique identifier for this tool call
	ID string `json:"id"`
	// Name is the name of the tool being called
	Name string `json:"name"`
	// Parameters are the arguments passed to the tool
	Parameters map[string]interface{} `json:"parameters"`
}

// ToolCall is an alias for ToolUseRequest for backward compatibility
type ToolCall = ToolUseRequest

// ImageData represents an image in a message
type ImageData struct {
	// URL is a URL pointing to the image
	URL string `json:"url,omitempty"`
	// Base64 is the base64-encoded image data
	Base64 string `json:"base64,omitempty"`
	// MimeType is the MIME type of the image (e.g., "image/png")
	MimeType string `json:"mime_type,omitempty"`
}

// MessageChunk represents a chunk of a streaming response
type MessageChunk struct {
	// Delta is the text content delta
	Delta string `json:"delta,omitempty"`
	// ToolCalls contains any completed tool calls
	ToolCalls []ToolUseRequest `json:"tool_calls,omitempty"`
	// FinishReason is set when the stream completes (e.g., "stop", "tool_use")
	FinishReason string `json:"finish_reason,omitempty"`
	// Error is set if an error occurred during streaming
	Error error `json:"error,omitempty"`
	// MessageID is the provider's ID for this message (set on final chunk)
	MessageID string `json:"message_id,omitempty"`
}

// ProviderConfig holds configuration for a cloud provider
type ProviderConfig struct {
	// APIKey is the authentication key for the provider
	APIKey string `json:"api_key,omitempty"`
	// BaseURL is an optional custom base URL for the provider's API
	BaseURL string `json:"base_url,omitempty"`
	// OrganizationID is an optional organization identifier
	OrganizationID string `json:"organization_id,omitempty"`
	// SessionKey is an optional session token (for browser-based auth)
	SessionKey string `json:"session_key,omitempty"`
	// Extra holds provider-specific configuration
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// ProviderCapabilities describes what a provider can do
type ProviderCapabilities struct {
	// SupportsTools indicates if the provider supports tool use
	SupportsTools bool `json:"supports_tools"`
	// SupportsVision indicates if the provider supports images
	SupportsVision bool `json:"supports_vision"`
	// SupportsStreaming indicates if the provider supports streaming responses
	SupportsStreaming bool `json:"supports_streaming"`
	// SupportedModels lists the available models
	SupportedModels []string `json:"supported_models,omitempty"`
	// MaxContextWindow is the maximum context size in tokens
	MaxContextWindow int `json:"max_context_window,omitempty"`
}
