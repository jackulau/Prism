package providers

import (
	"context"
	"errors"
	"testing"
)

// mockProvider is a mock implementation of AgentProvider for testing
type mockProvider struct {
	name         string
	agents       map[string]*Agent
	capabilities ProviderCapabilities
}

func newMockProvider(name string) *mockProvider {
	return &mockProvider{
		name:   name,
		agents: make(map[string]*Agent),
		capabilities: ProviderCapabilities{
			Streaming: true,
			Tools:     true,
		},
	}
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) CreateAgent(ctx context.Context, req CreateAgentRequest) (*Agent, error) {
	agent := &Agent{
		ID:           "test-agent-id",
		ProviderName: m.name,
		UserID:       req.UserID,
		Name:         req.Name,
		Status:       AgentStatusIdle,
		LLMProvider:  req.Provider,
		Model:        req.Model,
	}
	m.agents[agent.ID] = agent
	return agent, nil
}

func (m *mockProvider) GetAgent(ctx context.Context, agentID string) (*Agent, error) {
	agent, ok := m.agents[agentID]
	if !ok {
		return nil, errors.New("agent not found")
	}
	return agent, nil
}

func (m *mockProvider) SendMessage(ctx context.Context, agentID string, message string) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	go func() {
		defer close(ch)
		ch <- StreamChunk{
			Type:  StreamChunkTypeText,
			Delta: "Hello from mock provider!",
		}
		ch <- StreamChunk{
			Type: StreamChunkTypeDone,
			Done: true,
		}
	}()
	return ch, nil
}

func (m *mockProvider) GetMessages(ctx context.Context, agentID string) ([]Message, error) {
	return []Message{}, nil
}

func (m *mockProvider) StopAgent(ctx context.Context, agentID string) error {
	return nil
}

func (m *mockProvider) SupportsStreaming() bool {
	return true
}

func (m *mockProvider) Capabilities() ProviderCapabilities {
	return m.capabilities
}

func TestManagerRegisterProvider(t *testing.T) {
	m := NewManager()

	provider := newMockProvider("test-provider")
	err := m.RegisterProvider(provider)
	if err != nil {
		t.Fatalf("failed to register provider: %v", err)
	}

	// Registering same provider again should fail
	err = m.RegisterProvider(provider)
	if err != ErrProviderAlreadyExists {
		t.Errorf("expected ErrProviderAlreadyExists, got %v", err)
	}
}

func TestManagerGetProvider(t *testing.T) {
	m := NewManager()

	provider := newMockProvider("test-provider")
	_ = m.RegisterProvider(provider)

	// Get existing provider
	p, err := m.GetProvider("test-provider")
	if err != nil {
		t.Fatalf("failed to get provider: %v", err)
	}
	if p.Name() != "test-provider" {
		t.Errorf("expected provider name 'test-provider', got '%s'", p.Name())
	}

	// Get non-existent provider
	_, err = m.GetProvider("non-existent")
	if err != ErrProviderNotFound {
		t.Errorf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestManagerUnregisterProvider(t *testing.T) {
	m := NewManager()

	provider := newMockProvider("test-provider")
	_ = m.RegisterProvider(provider)

	// Unregister existing provider
	err := m.UnregisterProvider("test-provider")
	if err != nil {
		t.Fatalf("failed to unregister provider: %v", err)
	}

	// Provider should no longer exist
	_, err = m.GetProvider("test-provider")
	if err != ErrProviderNotFound {
		t.Errorf("expected ErrProviderNotFound after unregister, got %v", err)
	}

	// Unregister non-existent provider should fail
	err = m.UnregisterProvider("non-existent")
	if err != ErrProviderNotFound {
		t.Errorf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestManagerListProviders(t *testing.T) {
	m := NewManager()

	provider1 := newMockProvider("provider1")
	provider2 := newMockProvider("provider2")

	_ = m.RegisterProvider(provider1)
	_ = m.RegisterProvider(provider2)

	summaries := m.ListProviders()

	if len(summaries) != 2 {
		t.Errorf("expected 2 providers, got %d", len(summaries))
	}

	names := make(map[string]bool)
	for _, s := range summaries {
		names[s.Name] = true
	}

	if !names["provider1"] || !names["provider2"] {
		t.Error("expected both providers in list")
	}
}

func TestManagerCreateAgent(t *testing.T) {
	m := NewManager()
	m.Start()
	defer m.Stop()

	provider := newMockProvider("test-provider")
	_ = m.RegisterProvider(provider)

	ctx := context.Background()
	agent, err := m.CreateAgent(ctx, "test-provider", CreateAgentRequest{
		UserID:   "user-1",
		Name:     "Test Agent",
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5-20250929",
	})

	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	if agent.Name != "Test Agent" {
		t.Errorf("expected agent name 'Test Agent', got '%s'", agent.Name)
	}

	if agent.ProviderName != "test-provider" {
		t.Errorf("expected provider name 'test-provider', got '%s'", agent.ProviderName)
	}
}

func TestManagerCreateAgentNotStarted(t *testing.T) {
	m := NewManager()
	// Don't start the manager

	provider := newMockProvider("test-provider")
	_ = m.RegisterProvider(provider)

	ctx := context.Background()
	_, err := m.CreateAgent(ctx, "test-provider", CreateAgentRequest{})

	if err != ErrManagerNotStarted {
		t.Errorf("expected ErrManagerNotStarted, got %v", err)
	}
}

func TestManagerSendMessage(t *testing.T) {
	m := NewManager()
	m.Start()
	defer m.Stop()

	provider := newMockProvider("test-provider")
	_ = m.RegisterProvider(provider)

	ctx := context.Background()

	// Create an agent first
	agent, _ := m.CreateAgent(ctx, "test-provider", CreateAgentRequest{
		UserID: "user-1",
	})

	// Send a message
	ch, err := m.SendMessage(ctx, "test-provider", agent.ID, "Hello!")
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	// Read the response
	var chunks []StreamChunk
	for chunk := range ch {
		chunks = append(chunks, chunk)
	}

	if len(chunks) < 2 {
		t.Errorf("expected at least 2 chunks, got %d", len(chunks))
	}

	if chunks[0].Delta != "Hello from mock provider!" {
		t.Errorf("expected 'Hello from mock provider!', got '%s'", chunks[0].Delta)
	}
}

func TestManagerIsProviderAvailable(t *testing.T) {
	m := NewManager()

	provider := newMockProvider("test-provider")
	_ = m.RegisterProvider(provider)

	if !m.IsProviderAvailable("test-provider") {
		t.Error("expected test-provider to be available")
	}

	if m.IsProviderAvailable("non-existent") {
		t.Error("expected non-existent provider to not be available")
	}
}

func TestManagerDefaultProvider(t *testing.T) {
	m := NewManager()

	defaultProvider := m.DefaultProvider()
	if defaultProvider != "prism" {
		t.Errorf("expected default provider 'prism', got '%s'", defaultProvider)
	}
}

func TestManagerGetProviderCapabilities(t *testing.T) {
	m := NewManager()

	provider := newMockProvider("test-provider")
	provider.capabilities = ProviderCapabilities{
		Streaming:    true,
		Tools:        true,
		Vision:       false,
		CostTracking: true,
	}
	_ = m.RegisterProvider(provider)

	caps, err := m.GetProviderCapabilities("test-provider")
	if err != nil {
		t.Fatalf("failed to get capabilities: %v", err)
	}

	if !caps.Streaming {
		t.Error("expected streaming capability to be true")
	}
	if !caps.Tools {
		t.Error("expected tools capability to be true")
	}
	if caps.Vision {
		t.Error("expected vision capability to be false")
	}
	if !caps.CostTracking {
		t.Error("expected cost tracking capability to be true")
	}
}
