package providers

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages agent providers
type Manager struct {
	providers map[string]AgentProvider
	mu        sync.RWMutex
}

// NewManager creates a new agent provider manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]AgentProvider),
	}
}

// RegisterProvider registers a provider
func (m *Manager) RegisterProvider(provider AgentProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.Name()] = provider
}

// GetProvider gets a provider by name
func (m *Manager) GetProvider(name string) (AgentProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("agent provider not found: %s", name)
	}
	return provider, nil
}

// ListProviders returns all registered provider names
func (m *Manager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}

// HasProvider checks if a provider is registered
func (m *Manager) HasProvider(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.providers[name]
	return ok
}

// CreateAgent creates an agent using the specified provider
func (m *Manager) CreateAgent(ctx context.Context, providerName string, req CreateAgentRequest) (*Agent, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return provider.CreateAgent(ctx, req)
}

// GetAgent retrieves an agent using the specified provider
func (m *Manager) GetAgent(ctx context.Context, providerName, agentID string) (*Agent, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return provider.GetAgent(ctx, agentID)
}

// SendMessage sends a message to an agent using the specified provider
func (m *Manager) SendMessage(ctx context.Context, providerName, agentID, message string) (<-chan StreamChunk, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return provider.SendMessage(ctx, agentID, message)
}

// GetMessages retrieves messages from an agent using the specified provider
func (m *Manager) GetMessages(ctx context.Context, providerName, agentID string) ([]Message, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return provider.GetMessages(ctx, agentID)
}

// StopAgent stops an agent using the specified provider
func (m *Manager) StopAgent(ctx context.Context, providerName, agentID string) error {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return err
	}
	return provider.StopAgent(ctx, agentID)
}

// ProviderInfo contains information about an agent provider
type ProviderInfo struct {
	Name              string `json:"name"`
	SupportsStreaming bool   `json:"supports_streaming"`
}

// GetProviderInfo returns information about all registered providers
func (m *Manager) GetProviderInfo() []ProviderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]ProviderInfo, 0, len(m.providers))
	for _, provider := range m.providers {
		infos = append(infos, ProviderInfo{
			Name:              provider.Name(),
			SupportsStreaming: provider.SupportsStreaming(),
		})
	}
	return infos
}
