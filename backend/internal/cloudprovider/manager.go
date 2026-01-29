package cloudprovider

import (
	"context"
	"sync"
)

// ProviderManager is the concrete implementation of the Manager interface.
// It provides thread-safe registration and access to cloud providers.
type ProviderManager struct {
	providers map[string]CloudProvider
	mu        sync.RWMutex
}

// NewManager creates a new ProviderManager.
func NewManager() *ProviderManager {
	return &ProviderManager{
		providers: make(map[string]CloudProvider),
	}
}

// RegisterProvider registers a provider with the manager.
// If a provider with the same name is already registered, it will be replaced.
func (m *ProviderManager) RegisterProvider(provider CloudProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.Name()] = provider
}

// GetProvider returns a provider by name.
// Returns ErrProviderNotFound if the provider is not registered.
func (m *ProviderManager) GetProvider(name string) (CloudProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[name]
	if !ok {
		return nil, ErrProviderNotFound
	}
	return provider, nil
}

// ListProviders returns information about all registered providers.
func (m *ProviderManager) ListProviders() []ProviderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]ProviderInfo, 0, len(m.providers))
	for _, provider := range m.providers {
		infos = append(infos, ProviderInfo{
			Name:           provider.Name(),
			Capabilities:   provider.Capabilities(),
			HasCredentials: provider.HasCredentials(),
		})
	}
	return infos
}

// ListProviderNames returns the names of all registered providers.
func (m *ProviderManager) ListProviderNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}

// HasProvider returns true if a provider with the given name is registered.
func (m *ProviderManager) HasProvider(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.providers[name]
	return ok
}

// CreateAgent creates an agent using the specified provider.
func (m *ProviderManager) CreateAgent(ctx context.Context, providerName string, params CreateAgentParams) (*Agent, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return provider.CreateAgent(ctx, params)
}

// GetAgent retrieves an agent by its provider ID using the specified provider.
func (m *ProviderManager) GetAgent(ctx context.Context, providerName string, agentID string) (*Agent, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return provider.GetAgent(ctx, agentID)
}

// DeleteAgent removes an agent using the specified provider.
func (m *ProviderManager) DeleteAgent(ctx context.Context, providerName string, agentID string) error {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return err
	}
	return provider.DeleteAgent(ctx, agentID)
}

// GetMessages retrieves messages for an agent using its provider.
func (m *ProviderManager) GetMessages(ctx context.Context, agent *Agent) ([]ProviderMessage, error) {
	provider, err := m.GetProvider(agent.Provider)
	if err != nil {
		return nil, err
	}
	return provider.GetMessages(ctx, agent)
}

// SendMessage sends a message to an agent using its provider.
func (m *ProviderManager) SendMessage(ctx context.Context, agent *Agent, message string, images []ImageData) (bool, error) {
	provider, err := m.GetProvider(agent.Provider)
	if err != nil {
		return false, err
	}
	return provider.SendMessage(ctx, agent, message, images)
}

// StreamMessages streams messages from an agent using its provider.
func (m *ProviderManager) StreamMessages(ctx context.Context, agent *Agent) (<-chan MessageChunk, error) {
	provider, err := m.GetProvider(agent.Provider)
	if err != nil {
		return nil, err
	}
	return provider.StreamMessages(ctx, agent)
}

// ValidateCredentials validates credentials for the specified provider.
func (m *ProviderManager) ValidateCredentials(ctx context.Context, providerName string) error {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return err
	}
	return provider.ValidateCredentials(ctx)
}

// Ensure ProviderManager implements Manager interface.
var _ Manager = (*ProviderManager)(nil)
