---
id: jules-provider
name: Jules Provider - External AI Integration
wave: 2
priority: 2
dependencies:
- prism-provider
estimated_hours: 4
tags:
- backend
- providers
- external-api
---

## Objective

Implement a Jules AI provider that integrates with the external Jules API to create and manage agents through their service.

## Context

Jules provides an AI agent API similar to Cursor. This provider will implement the same `AgentProvider` interface as the Prism and Cursor providers but route requests to Jules' external API. The implementation pattern should closely follow the Cursor provider structure.

## API Reference

### Assumed Endpoints (to be confirmed from Jules documentation)
- `POST https://api.jules.ai/v1/agents` - Create agent
- `GET https://api.jules.ai/v1/agents/{id}` - Get agent status
- `GET https://api.jules.ai/v1/agents/{id}/messages` - Get message history
- `POST https://api.jules.ai/v1/agents/{id}/messages` - Send message

### Authentication
- Bearer token or API key auth (confirm from Jules docs)
- Header: `Authorization: Bearer {api_key}` (assumed)

## Implementation

### 1. Create Jules Provider

Create `backend/internal/providers/jules/provider.go`:
```go
type JulesProvider struct {
    client    *http.Client
    baseURL   string
    apiKey    string
}

func (p *JulesProvider) Name() string { return "jules" }

func (p *JulesProvider) CreateAgent(ctx context.Context, req CreateAgentRequest) (*Agent, error) {
    // POST to /v1/agents
}

func (p *JulesProvider) GetAgent(ctx context.Context, agentID string) (*Agent, error) {
    // GET /v1/agents/{id}
}

func (p *JulesProvider) SendMessage(ctx context.Context, agentID, message string) (<-chan StreamChunk, error) {
    // POST to /v1/agents/{id}/messages with streaming
}

func (p *JulesProvider) GetMessages(ctx context.Context, agentID string) ([]Message, error) {
    // GET /v1/agents/{id}/messages
}
```

### 2. HTTP Client Setup

Create `backend/internal/providers/jules/client.go`:
- Configure HTTP client with timeouts
- Handle authentication (Bearer token or API key)
- Parse responses (SSE if streaming supported)
- Error handling for API errors

### 3. Request/Response Types

Create `backend/internal/providers/jules/types.go`:
```go
type CreateAgentRequest struct {
    Prompt      string            `json:"prompt"`
    Model       string            `json:"model,omitempty"`
    Repository  string            `json:"repository,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

type AgentResponse struct {
    ID        string `json:"id"`
    Status    string `json:"status"`
    Branch    string `json:"branch,omitempty"`
    CreatedAt string `json:"created_at"`
}

type MessageResponse struct {
    ID        string `json:"id"`
    Role      string `json:"role"`
    Content   string `json:"content"`
    Timestamp string `json:"timestamp"`
}
```

### 4. Streaming Support

- Parse streaming responses (if Jules supports SSE)
- Fall back to polling if no streaming
- Convert to internal `StreamChunk` format

### 5. API Key Management

- Store encrypted in `provider_keys` table
- Use existing encryption service
- Provider-specific key validation

### 6. Jules-Specific Features

- Handle repository context if Jules supports it
- Branch creation/management if supported
- GitHub integration if applicable

## Files to Create/Modify

- `backend/internal/providers/jules/provider.go` - Main provider implementation
- `backend/internal/providers/jules/client.go` - HTTP client utilities
- `backend/internal/providers/jules/types.go` - Request/response types
- `backend/internal/providers/jules/streaming.go` - SSE/streaming parsing
- `backend/internal/providers/manager.go` - Register Jules provider

## Acceptance Criteria

- [ ] JulesProvider implements AgentProvider interface
- [ ] Authentication with Jules API works correctly
- [ ] Can create agents via Jules API
- [ ] Can retrieve agent status
- [ ] Can retrieve message history
- [ ] Can send messages (with streaming if supported)
- [ ] API errors handled gracefully
- [ ] API key stored encrypted in database
- [ ] Provider registered with provider manager
- [ ] Unit tests with mocked HTTP responses

## Notes

- Jules API documentation should be consulted for exact endpoints
- This task can be developed in parallel with cursor-provider
- Both external providers follow the same pattern

## Integration Points

- **Provides**: Jules agent creation/messaging capability
- **Consumes**: AgentProvider interface from prism-provider task
- **Conflicts**: None - isolated external API client
