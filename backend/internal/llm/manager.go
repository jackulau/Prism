package llm

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages LLM providers
type Manager struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewManager creates a new LLM manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]Provider),
	}
}

// RegisterProvider registers a provider
func (m *Manager) RegisterProvider(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.Name()] = provider
}

// GetProvider gets a provider by name
func (m *Manager) GetProvider(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// ListProviders returns all registered providers
func (m *Manager) ListProviders() []ProviderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var infos []ProviderInfo
	for name, provider := range m.providers {
		infos = append(infos, ProviderInfo{
			Name:          name,
			Models:        provider.Models(),
			SupportsTools: provider.SupportsTools(),
			SupportsVision: provider.SupportsVision(),
		})
	}
	return infos
}

// Chat sends a chat request to the appropriate provider
func (m *Manager) Chat(ctx context.Context, providerName string, req *ChatRequest) (<-chan StreamChunk, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return provider.Chat(ctx, req)
}

// ProviderInfo contains information about a provider
type ProviderInfo struct {
	Name           string  `json:"name"`
	Models         []Model `json:"models"`
	SupportsTools  bool    `json:"supports_tools"`
	SupportsVision bool    `json:"supports_vision"`
}
