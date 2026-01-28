package sandbox

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages multiple sandbox providers
type Manager struct {
	providers      map[string]Provider
	defaultProvider string
	mu             sync.RWMutex
}

// NewManager creates a new sandbox manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]Provider),
	}
}

// RegisterProvider registers a sandbox provider
func (m *Manager) RegisterProvider(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := provider.Name()
	m.providers[name] = provider

	// Set as default if it's the first provider
	if m.defaultProvider == "" {
		m.defaultProvider = name
	}
}

// SetDefault sets the default provider by name
func (m *Manager) SetDefault(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.providers[name]; !ok {
		return fmt.Errorf("provider not found: %s", name)
	}

	m.defaultProvider = name
	return nil
}

// GetProvider returns a provider by name
func (m *Manager) GetProvider(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return provider, nil
}

// GetDefaultProvider returns the default provider
func (m *Manager) GetDefaultProvider() (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.defaultProvider == "" {
		return nil, fmt.Errorf("no default provider configured")
	}

	provider, ok := m.providers[m.defaultProvider]
	if !ok {
		return nil, fmt.Errorf("default provider not found: %s", m.defaultProvider)
	}

	return provider, nil
}

// ListProviders returns a list of registered provider names
func (m *Manager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}

// CreateSandbox creates a sandbox using the specified provider (or default if empty)
func (m *Manager) CreateSandbox(ctx context.Context, providerName string, opts *CreateOptions) (*Sandbox, error) {
	var provider Provider
	var err error

	if providerName == "" {
		provider, err = m.GetDefaultProvider()
	} else {
		provider, err = m.GetProvider(providerName)
	}

	if err != nil {
		return nil, err
	}

	return provider.CreateSandbox(ctx, opts)
}

// DeploySandbox deploys files to a sandbox using the specified provider
func (m *Manager) DeploySandbox(ctx context.Context, providerName, sandboxID string, files map[string][]byte) (*DeployResult, error) {
	var provider Provider
	var err error

	if providerName == "" {
		provider, err = m.GetDefaultProvider()
	} else {
		provider, err = m.GetProvider(providerName)
	}

	if err != nil {
		return nil, err
	}

	return provider.DeploySandbox(ctx, sandboxID, files)
}

// GetSandbox retrieves a sandbox from the specified provider
func (m *Manager) GetSandbox(ctx context.Context, providerName, sandboxID string) (*Sandbox, error) {
	var provider Provider
	var err error

	if providerName == "" {
		provider, err = m.GetDefaultProvider()
	} else {
		provider, err = m.GetProvider(providerName)
	}

	if err != nil {
		return nil, err
	}

	return provider.GetSandbox(ctx, sandboxID)
}

// GetLogs retrieves logs from the specified provider
func (m *Manager) GetLogs(ctx context.Context, providerName, sandboxID string) (<-chan LogEntry, error) {
	var provider Provider
	var err error

	if providerName == "" {
		provider, err = m.GetDefaultProvider()
	} else {
		provider, err = m.GetProvider(providerName)
	}

	if err != nil {
		return nil, err
	}

	return provider.GetLogs(ctx, sandboxID)
}

// DeleteSandbox deletes a sandbox from the specified provider
func (m *Manager) DeleteSandbox(ctx context.Context, providerName, sandboxID string) error {
	var provider Provider
	var err error

	if providerName == "" {
		provider, err = m.GetDefaultProvider()
	} else {
		provider, err = m.GetProvider(providerName)
	}

	if err != nil {
		return err
	}

	return provider.DeleteSandbox(ctx, sandboxID)
}

// GetPreviewURL gets the preview URL from the specified provider
func (m *Manager) GetPreviewURL(providerName, sandboxID string) (string, error) {
	var provider Provider
	var err error

	if providerName == "" {
		provider, err = m.GetDefaultProvider()
	} else {
		provider, err = m.GetProvider(providerName)
	}

	if err != nil {
		return "", err
	}

	return provider.GetPreviewURL(sandboxID), nil
}
