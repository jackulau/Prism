// Package cloudprovider defines interfaces for cloud-based AI agent backends.
//
// Unlike the llm.Provider interface which handles direct LLM chat via API calls,
// CloudProvider abstracts remote agent lifecycle management for cloud-hosted agents
// (such as Claude.ai conversations). This allows the application to create, manage,
// and communicate with AI agents running on cloud platforms.
package cloudprovider

import (
	"context"
)

// CloudProvider defines the interface for cloud-based agent backends.
// Implementations of this interface handle communication with cloud AI services
// that manage agent lifecycles, such as Claude.ai or OpenAI Assistants.
type CloudProvider interface {
	// Name returns the provider name (e.g., "claude-cloud", "openai-assistants").
	// This should be a unique identifier for the provider.
	Name() string

	// CreateAgent creates a new agent with the given parameters.
	// Returns the created Agent or an error if creation fails.
	CreateAgent(ctx context.Context, params CreateAgentParams) (*Agent, error)

	// GetAgent retrieves an agent by its provider ID.
	// Returns ErrAgentNotFound if the agent doesn't exist.
	GetAgent(ctx context.Context, agentID string) (*Agent, error)

	// DeleteAgent removes an agent from the cloud provider.
	// Returns ErrAgentNotFound if the agent doesn't exist.
	DeleteAgent(ctx context.Context, agentID string) error

	// GetMessages retrieves all messages from an agent's conversation history.
	// Returns messages in chronological order.
	GetMessages(ctx context.Context, agent *Agent) ([]ProviderMessage, error)

	// SendMessage sends a message to an agent.
	// The images parameter allows sending images along with the message.
	// Returns true if the message was sent successfully.
	SendMessage(ctx context.Context, agent *Agent, message string, images []ImageData) (bool, error)

	// StreamMessages returns a channel that streams the agent's response.
	// The channel will be closed when the response is complete or an error occurs.
	// Check MessageChunk.Error for any errors during streaming.
	StreamMessages(ctx context.Context, agent *Agent) (<-chan MessageChunk, error)

	// ValidateCredentials validates that the provider's credentials are valid.
	// This typically makes a lightweight API call to verify authentication.
	ValidateCredentials(ctx context.Context) error

	// HasCredentials returns whether credentials are configured for this provider.
	// This does not validate the credentials, only checks if they are present.
	HasCredentials() bool

	// Capabilities returns the capabilities of this provider.
	Capabilities() ProviderCapabilities
}

// Manager manages multiple cloud providers.
// It provides a centralized way to register and access cloud providers.
type Manager interface {
	// RegisterProvider registers a provider with the manager.
	RegisterProvider(provider CloudProvider)

	// GetProvider returns a provider by name.
	// Returns an error if the provider is not registered.
	GetProvider(name string) (CloudProvider, error)

	// ListProviders returns information about all registered providers.
	ListProviders() []ProviderInfo
}

// ProviderInfo contains summary information about a registered provider.
type ProviderInfo struct {
	// Name is the provider's unique identifier
	Name string `json:"name"`
	// Capabilities describes what the provider can do
	Capabilities ProviderCapabilities `json:"capabilities"`
	// HasCredentials indicates if credentials are configured
	HasCredentials bool `json:"has_credentials"`
}
