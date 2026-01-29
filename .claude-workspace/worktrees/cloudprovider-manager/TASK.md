---
id: cloudprovider-manager
name: CloudProvider Manager and Registry
wave: 2
priority: 2
dependencies:
- cloudprovider-interface
estimated_hours: 3
tags:
- backend
- manager
- registry
---

## Objective

Create a manager/registry for CloudProvider implementations, following the pattern established by `llm.Manager`.

## Context

The existing `backend/internal/llm/manager.go` provides a good template. The CloudProvider manager will:
- Register and manage cloud provider implementations
- Handle credential management per user
- Route requests to appropriate providers
- Provide discovery of available providers

## Implementation

1. **Create manager file**: `backend/internal/cloudprovider/manager.go`
   ```go
   package cloudprovider
   
   import (
       "context"
       "sync"
   )
   
   // Manager manages cloud provider registrations and routing
   type Manager struct {
       providers map[string]CloudProvider
       mu        sync.RWMutex
   }
   
   func NewManager() *Manager {
       return &Manager{
           providers: make(map[string]CloudProvider),
       }
   }
   
   // RegisterProvider adds a provider to the registry
   func (m *Manager) RegisterProvider(provider CloudProvider) {
       m.mu.Lock()
       defer m.mu.Unlock()
       m.providers[provider.Name()] = provider
   }
   
   // GetProvider retrieves a provider by name
   func (m *Manager) GetProvider(name string) (CloudProvider, error) {
       m.mu.RLock()
       defer m.mu.RUnlock()
       provider, ok := m.providers[name]
       if !ok {
           return nil, ErrProviderNotFound
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
   
   // CreateAgent creates an agent using the specified provider
   func (m *Manager) CreateAgent(ctx context.Context, providerName string, params CreateAgentParams) (*Agent, error) {
       provider, err := m.GetProvider(providerName)
       if err != nil {
           return nil, err
       }
       return provider.CreateAgent(ctx, params)
   }
   
   // GetMessages retrieves messages for an agent
   func (m *Manager) GetMessages(ctx context.Context, providerName string, agent *Agent) ([]ProviderMessage, error) {
       provider, err := m.GetProvider(providerName)
       if err != nil {
           return nil, err
       }
       return provider.GetMessages(ctx, agent)
   }
   
   // SendMessage sends a message to an agent
   func (m *Manager) SendMessage(ctx context.Context, providerName string, agent *Agent, message string, images []ImageData) (bool, error) {
       provider, err := m.GetProvider(providerName)
       if err != nil {
           return false, err
       }
       return provider.SendMessage(ctx, agent, message, images)
   }
   ```

2. **Add manager errors**: Update `backend/internal/cloudprovider/errors.go`
   - Add `ErrProviderNotFound`
   - Add `ErrProviderAlreadyRegistered`

3. **Wire into dependencies**: Update initialization pattern in `backend/cmd/server/main.go`
   - Create and configure CloudProvider manager
   - Register with Dependencies struct

## Acceptance Criteria

- [ ] Manager provides thread-safe provider registration
- [ ] Manager routes requests to correct providers
- [ ] Provider discovery returns all registered providers
- [ ] Manager follows same pattern as `llm.Manager`
- [ ] Error handling is consistent with codebase

## Files to Create/Modify

- `backend/internal/cloudprovider/manager.go` - Manager implementation
- `backend/internal/cloudprovider/errors.go` - Additional errors (modify)
- `backend/internal/api/routes/router.go` - Add to Dependencies struct
- `backend/cmd/server/main.go` - Initialize manager

## Integration Points

- **Provides**: Centralized cloud provider management
- **Consumes**: CloudProvider implementations, Dependencies struct
- **Conflicts**: Avoid editing `router.go` simultaneously with other tasks
