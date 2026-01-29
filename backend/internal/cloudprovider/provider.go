package cloudprovider

import (
	"context"
)

// CloudProvider defines the interface for cloud-based agent backends.
// This abstracts cloud-hosted agent services that create, manage, and
// communicate with AI agents (different from llm.Provider which handles
// direct LLM chat).
type CloudProvider interface {
	// Name returns the provider name (e.g., "anthropic-cloud", "openai-assistants")
	Name() string

	// CreateAgent creates a new agent with the given parameters
	CreateAgent(ctx context.Context, params CreateAgentParams) (*Agent, error)

	// GetAgent retrieves an agent by ID
	GetAgent(ctx context.Context, agentID string) (*Agent, error)

	// DeleteAgent removes an agent
	DeleteAgent(ctx context.Context, agentID string) error

	// GetMessages retrieves all messages for an agent's conversation
	GetMessages(ctx context.Context, agent *Agent) ([]ProviderMessage, error)

	// SendMessage sends a message to an agent and returns success status
	SendMessage(ctx context.Context, agent *Agent, message string, images []ImageData) (bool, error)

	// StreamMessages returns a channel for streaming agent responses
	StreamMessages(ctx context.Context, agent *Agent) (<-chan MessageChunk, error)

	// ValidateCredentials validates the provider credentials
	ValidateCredentials(ctx context.Context) error

	// HasCredentials returns whether credentials are configured
	HasCredentials() bool
}
