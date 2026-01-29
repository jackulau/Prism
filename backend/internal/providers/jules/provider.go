package jules

import (
	"context"
	"time"

	"github.com/jacklau/prism/internal/providers"
)

// Provider implements the AgentProvider interface for Jules API
type Provider struct {
	client *HTTPClient
}

// NewProvider creates a new Jules provider
func NewProvider(apiKey string) *Provider {
	return &Provider{
		client: NewHTTPClient(apiKey, ""),
	}
}

// NewProviderWithBaseURL creates a new Jules provider with a custom base URL
func NewProviderWithBaseURL(apiKey, baseURL string) *Provider {
	return &Provider{
		client: NewHTTPClient(apiKey, baseURL),
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "jules"
}

// CreateAgent creates a new agent via Jules API
func (p *Provider) CreateAgent(ctx context.Context, req providers.CreateAgentRequest) (*providers.Agent, error) {
	julesReq := JulesCreateAgentRequest{
		Prompt: req.Prompt,
		Model:  req.Model,
	}

	// Extract repository from metadata if present
	if req.Metadata != nil {
		if repo, ok := req.Metadata["repository"].(string); ok {
			julesReq.Repository = repo
		}
	}

	resp, err := p.client.CreateAgent(ctx, julesReq)
	if err != nil {
		return nil, err
	}

	createdAt := time.Now()
	if resp.CreatedAt != "" {
		if parsed, parseErr := time.Parse(time.RFC3339, resp.CreatedAt); parseErr == nil {
			createdAt = parsed
		}
	}

	metadata := map[string]interface{}{
		"model": resp.Model,
	}
	if resp.Branch != "" {
		metadata["branch"] = resp.Branch
	}

	return &providers.Agent{
		ID:        resp.ID,
		Provider:  "jules",
		Status:    mapJulesStatus(resp.Status),
		CreatedAt: createdAt,
		Metadata:  metadata,
	}, nil
}

// GetAgent retrieves an agent by ID
func (p *Provider) GetAgent(ctx context.Context, agentID string) (*providers.Agent, error) {
	resp, err := p.client.GetAgent(ctx, agentID)
	if err != nil {
		return nil, err
	}

	createdAt := time.Now()
	if resp.CreatedAt != "" {
		if parsed, parseErr := time.Parse(time.RFC3339, resp.CreatedAt); parseErr == nil {
			createdAt = parsed
		}
	}

	metadata := map[string]interface{}{
		"model": resp.Model,
	}
	if resp.Branch != "" {
		metadata["branch"] = resp.Branch
	}

	return &providers.Agent{
		ID:        resp.ID,
		Provider:  "jules",
		Status:    mapJulesStatus(resp.Status),
		CreatedAt: createdAt,
		Metadata:  metadata,
	}, nil
}

// SendMessage sends a message to an agent and returns a streaming response
func (p *Provider) SendMessage(ctx context.Context, agentID string, message string) (<-chan providers.StreamChunk, error) {
	resp, err := p.client.SendMessage(ctx, agentID, message)
	if err != nil {
		return nil, err
	}

	chunks := make(chan providers.StreamChunk, 100)

	go func() {
		defer resp.Body.Close()

		parser := NewStreamParser(resp.Body)
		parser.ParseStream(chunks)
	}()

	return chunks, nil
}

// GetMessages retrieves the message history for an agent
func (p *Provider) GetMessages(ctx context.Context, agentID string) ([]providers.Message, error) {
	resp, err := p.client.GetMessages(ctx, agentID)
	if err != nil {
		return nil, err
	}

	messages := make([]providers.Message, len(resp.Messages))
	for i, msg := range resp.Messages {
		createdAt := msg.CreatedAt
		if createdAt.IsZero() && msg.Timestamp != "" {
			if parsed, parseErr := time.Parse(time.RFC3339, msg.Timestamp); parseErr == nil {
				createdAt = parsed
			}
		}

		messages[i] = providers.Message{
			ID:        msg.ID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: createdAt,
		}
	}

	return messages, nil
}

// StopAgent stops a running agent
func (p *Provider) StopAgent(ctx context.Context, agentID string) error {
	return p.client.StopAgent(ctx, agentID)
}

// SupportsStreaming returns whether the provider supports streaming responses
func (p *Provider) SupportsStreaming() bool {
	return true
}

// SetAPIKey updates the provider's API key
func (p *Provider) SetAPIKey(key string) {
	p.client.SetAPIKey(key)
}

// HasConfiguredKey returns whether the provider has an API key configured
func (p *Provider) HasConfiguredKey() bool {
	return p.client.HasAPIKey()
}

// ValidateKey validates an API key
func (p *Provider) ValidateKey(ctx context.Context, key string) error {
	return p.client.ValidateKey(ctx, key)
}

// mapJulesStatus maps Jules API status strings to internal AgentStatus
func mapJulesStatus(status string) providers.AgentStatus {
	switch status {
	case "pending", "queued", "initializing":
		return providers.AgentStatusPending
	case "running", "in_progress", "active", "working":
		return providers.AgentStatusRunning
	case "completed", "done", "finished", "success":
		return providers.AgentStatusCompleted
	case "failed", "error":
		return providers.AgentStatusFailed
	case "stopped", "cancelled", "canceled", "aborted":
		return providers.AgentStatusStopped
	default:
		return providers.AgentStatusPending
	}
}
