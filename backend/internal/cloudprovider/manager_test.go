package cloudprovider

import (
	"context"
	"testing"
)

// mockProvider is a mock implementation of CloudProvider for testing.
type mockProvider struct {
	name           string
	hasCredentials bool
	capabilities   ProviderCapabilities
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) CreateAgent(ctx context.Context, params CreateAgentParams) (*Agent, error) {
	return &Agent{ID: "test-agent", Provider: m.name}, nil
}

func (m *mockProvider) GetAgent(ctx context.Context, agentID string) (*Agent, error) {
	return &Agent{ID: agentID, Provider: m.name}, nil
}

func (m *mockProvider) DeleteAgent(ctx context.Context, agentID string) error {
	return nil
}

func (m *mockProvider) GetMessages(ctx context.Context, agent *Agent) ([]ProviderMessage, error) {
	return []ProviderMessage{}, nil
}

func (m *mockProvider) SendMessage(ctx context.Context, agent *Agent, message string, images []ImageData) (bool, error) {
	return true, nil
}

func (m *mockProvider) StreamMessages(ctx context.Context, agent *Agent) (<-chan MessageChunk, error) {
	ch := make(chan MessageChunk)
	close(ch)
	return ch, nil
}

func (m *mockProvider) ValidateCredentials(ctx context.Context) error {
	return nil
}

func (m *mockProvider) HasCredentials() bool {
	return m.hasCredentials
}

func (m *mockProvider) Capabilities() ProviderCapabilities {
	return m.capabilities
}

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.providers == nil {
		t.Fatal("providers map not initialized")
	}
}

func TestRegisterProvider(t *testing.T) {
	manager := NewManager()
	provider := &mockProvider{name: "test-provider"}

	manager.RegisterProvider(provider)

	if !manager.HasProvider("test-provider") {
		t.Error("Provider was not registered")
	}
}

func TestGetProvider(t *testing.T) {
	manager := NewManager()
	provider := &mockProvider{name: "test-provider"}
	manager.RegisterProvider(provider)

	got, err := manager.GetProvider("test-provider")
	if err != nil {
		t.Fatalf("GetProvider returned error: %v", err)
	}
	if got.Name() != "test-provider" {
		t.Errorf("GetProvider returned wrong provider: got %s, want test-provider", got.Name())
	}
}

func TestGetProviderNotFound(t *testing.T) {
	manager := NewManager()

	_, err := manager.GetProvider("nonexistent")
	if err != ErrProviderNotFound {
		t.Errorf("Expected ErrProviderNotFound, got: %v", err)
	}
}

func TestListProviders(t *testing.T) {
	manager := NewManager()
	manager.RegisterProvider(&mockProvider{
		name:           "provider-1",
		hasCredentials: true,
		capabilities:   ProviderCapabilities{SupportsTools: true},
	})
	manager.RegisterProvider(&mockProvider{
		name:           "provider-2",
		hasCredentials: false,
		capabilities:   ProviderCapabilities{SupportsVision: true},
	})

	infos := manager.ListProviders()
	if len(infos) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(infos))
	}

	// Check that both providers are listed
	providerMap := make(map[string]ProviderInfo)
	for _, info := range infos {
		providerMap[info.Name] = info
	}

	if _, ok := providerMap["provider-1"]; !ok {
		t.Error("provider-1 not found in list")
	}
	if _, ok := providerMap["provider-2"]; !ok {
		t.Error("provider-2 not found in list")
	}

	// Check credentials and capabilities
	if !providerMap["provider-1"].HasCredentials {
		t.Error("provider-1 should have credentials")
	}
	if providerMap["provider-2"].HasCredentials {
		t.Error("provider-2 should not have credentials")
	}
}

func TestListProviderNames(t *testing.T) {
	manager := NewManager()
	manager.RegisterProvider(&mockProvider{name: "provider-a"})
	manager.RegisterProvider(&mockProvider{name: "provider-b"})

	names := manager.ListProviderNames()
	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}
}

func TestHasProvider(t *testing.T) {
	manager := NewManager()
	manager.RegisterProvider(&mockProvider{name: "exists"})

	if !manager.HasProvider("exists") {
		t.Error("HasProvider returned false for registered provider")
	}
	if manager.HasProvider("does-not-exist") {
		t.Error("HasProvider returned true for unregistered provider")
	}
}

func TestCreateAgent(t *testing.T) {
	manager := NewManager()
	manager.RegisterProvider(&mockProvider{name: "test-provider"})

	agent, err := manager.CreateAgent(context.Background(), "test-provider", CreateAgentParams{Name: "test"})
	if err != nil {
		t.Fatalf("CreateAgent returned error: %v", err)
	}
	if agent == nil {
		t.Fatal("CreateAgent returned nil agent")
	}
}

func TestCreateAgentProviderNotFound(t *testing.T) {
	manager := NewManager()

	_, err := manager.CreateAgent(context.Background(), "nonexistent", CreateAgentParams{})
	if err != ErrProviderNotFound {
		t.Errorf("Expected ErrProviderNotFound, got: %v", err)
	}
}

func TestSendMessage(t *testing.T) {
	manager := NewManager()
	manager.RegisterProvider(&mockProvider{name: "test-provider"})

	agent := &Agent{Provider: "test-provider"}
	ok, err := manager.SendMessage(context.Background(), agent, "hello", nil)
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	if !ok {
		t.Error("SendMessage returned false")
	}
}

func TestManagerImplementsInterface(t *testing.T) {
	// Compile-time check that ProviderManager implements Manager
	var _ Manager = (*ProviderManager)(nil)
}
