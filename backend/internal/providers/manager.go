package providers

import (
	"context"
	"errors"
	"sync"
)

// Manager errors
var (
	ErrProviderNotFound      = errors.New("provider not found")
	ErrProviderAlreadyExists = errors.New("provider already exists")
	ErrManagerNotStarted     = errors.New("manager not started")
)

// Manager manages multiple agent providers and routes requests to the appropriate one
type Manager struct {
	providers map[string]AgentProvider
	mu        sync.RWMutex

	// State
	running bool
}

// NewManager creates a new provider manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]AgentProvider),
	}
}

// RegisterProvider registers an agent provider
func (m *Manager) RegisterProvider(provider AgentProvider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := provider.Name()
	if _, exists := m.providers[name]; exists {
		return ErrProviderAlreadyExists
	}

	m.providers[name] = provider
	return nil
}

// UnregisterProvider removes a provider
func (m *Manager) UnregisterProvider(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.providers[name]; !exists {
		return ErrProviderNotFound
	}

	delete(m.providers, name)
	return nil
}

// GetProvider gets a provider by name
func (m *Manager) GetProvider(name string) (AgentProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[name]
	if !ok {
		return nil, ErrProviderNotFound
	}
	return provider, nil
}

// ListProviders returns all registered providers with their capabilities
func (m *Manager) ListProviders() []ProviderSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summaries := make([]ProviderSummary, 0, len(m.providers))
	for _, provider := range m.providers {
		summaries = append(summaries, ProviderSummary{
			Name:         provider.Name(),
			Capabilities: provider.Capabilities(),
		})
	}
	return summaries
}

// ProviderSummary contains summary information about a provider
type ProviderSummary struct {
	Name         string               `json:"name"`
	Capabilities ProviderCapabilities `json:"capabilities"`
}

// Start starts the manager and all registered providers
func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return
	}

	m.running = true

	// Start providers that have a Start method
	for _, provider := range m.providers {
		if starter, ok := provider.(interface{ Start() }); ok {
			starter.Start()
		}
	}
}

// Stop stops the manager and all registered providers
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.running = false

	// Stop providers that have a Stop method
	for _, provider := range m.providers {
		if stopper, ok := provider.(interface{ Stop() }); ok {
			stopper.Stop()
		}
	}
}

// CreateAgent creates an agent using the specified provider
func (m *Manager) CreateAgent(ctx context.Context, providerName string, req CreateAgentRequest) (*Agent, error) {
	m.mu.RLock()
	if !m.running {
		m.mu.RUnlock()
		return nil, ErrManagerNotStarted
	}
	m.mu.RUnlock()

	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.CreateAgent(ctx, req)
}

// GetAgent retrieves an agent by ID from the specified provider
func (m *Manager) GetAgent(ctx context.Context, providerName, agentID string) (*Agent, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.GetAgent(ctx, agentID)
}

// SendMessage sends a message to an agent
func (m *Manager) SendMessage(ctx context.Context, providerName, agentID, message string) (<-chan StreamChunk, error) {
	m.mu.RLock()
	if !m.running {
		m.mu.RUnlock()
		return nil, ErrManagerNotStarted
	}
	m.mu.RUnlock()

	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.SendMessage(ctx, agentID, message)
}

// GetMessages retrieves the message history for an agent
func (m *Manager) GetMessages(ctx context.Context, providerName, agentID string) ([]Message, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.GetMessages(ctx, agentID)
}

// StopAgent stops a running agent
func (m *Manager) StopAgent(ctx context.Context, providerName, agentID string) error {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return err
	}

	return provider.StopAgent(ctx, agentID)
}

// FindAgentProvider searches all providers to find which one has the given agent
// Returns the provider name and the agent if found
func (m *Manager) FindAgentProvider(ctx context.Context, agentID string) (string, *Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, provider := range m.providers {
		agent, err := provider.GetAgent(ctx, agentID)
		if err == nil && agent != nil {
			return name, agent, nil
		}
	}

	return "", nil, ErrProviderNotFound
}

// GetProviderCapabilities returns the capabilities for a specific provider
func (m *Manager) GetProviderCapabilities(providerName string) (*ProviderCapabilities, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	caps := provider.Capabilities()
	return &caps, nil
}

// IsProviderAvailable checks if a provider is registered
func (m *Manager) IsProviderAvailable(providerName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.providers[providerName]
	return exists
}

// DefaultProvider returns the default provider name
// For now, this returns "prism" as it's the native provider
func (m *Manager) DefaultProvider() string {
	return "prism"
}

// ExtendedProvider interface for providers with additional functionality
type ExtendedProvider interface {
	AgentProvider

	// ListAgents returns all agents for a user
	ListAgents(ctx context.Context, userID string) ([]*Agent, error)

	// DeleteAgent deletes an agent
	DeleteAgent(ctx context.Context, agentID string) error
}

// ListAgents returns all agents for a user from a specific provider
// Only works if the provider implements ExtendedProvider
func (m *Manager) ListAgents(ctx context.Context, providerName, userID string) ([]*Agent, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	extended, ok := provider.(ExtendedProvider)
	if !ok {
		return nil, errors.New("provider does not support listing agents")
	}

	return extended.ListAgents(ctx, userID)
}

// DeleteAgent deletes an agent from a specific provider
// Only works if the provider implements ExtendedProvider
func (m *Manager) DeleteAgent(ctx context.Context, providerName, agentID string) error {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return err
	}

	extended, ok := provider.(ExtendedProvider)
	if !ok {
		return errors.New("provider does not support deleting agents")
	}

	return extended.DeleteAgent(ctx, agentID)
}

// ListAllAgents returns all agents for a user across all providers
func (m *Manager) ListAllAgents(ctx context.Context, userID string) ([]*Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allAgents := make([]*Agent, 0)
	for _, provider := range m.providers {
		extended, ok := provider.(ExtendedProvider)
		if !ok {
			continue
		}

		agents, err := extended.ListAgents(ctx, userID)
		if err != nil {
			continue
		}

		allAgents = append(allAgents, agents...)
	}

	return allAgents, nil
}
