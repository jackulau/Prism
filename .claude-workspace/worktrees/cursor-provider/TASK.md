---
id: cursor-provider
name: Cursor Provider - External AI Integration
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

Implement a Cursor AI provider that integrates with the external Cursor API to create and manage agents through their service.

## Context

Cursor provides an AI agent API that allows creating agents, getting message history, and sending follow-up messages. This provider will implement the same `AgentProvider` interface as the Prism provider but route requests to Cursor's external API. Authentication uses Basic auth with the API key Base64 encoded.

## API Reference

### Endpoints
- `POST https://api.cursor.com/v0/agents` - Create agent
- `GET https://api.cursor.com/v0/agents/{id}/messages` - Get message history
- `POST https://api.cursor.com/v0/agents/{id}/followup` - Send follow-up message

### Authentication
- Basic auth with API key
- Header: `Authorization: Basic {base64(api_key + ":")}`

## Implementation

### 1. Create Cursor Provider

Create `backend/internal/providers/cursor/provider.go`:
```go
type CursorProvider struct {
    client    *http.Client
    baseURL   string
    apiKey    string
}

func (p *CursorProvider) Name() string { return "cursor" }

func (p *CursorProvider) CreateAgent(ctx context.Context, req CreateAgentRequest) (*Agent, error) {
    // POST to /v0/agents
}

func (p *CursorProvider) GetAgent(ctx context.Context, agentID string) (*Agent, error) {
    // GET agent status
}

func (p *CursorProvider) SendMessage(ctx context.Context, agentID, message string) (<-chan StreamChunk, error) {
    // POST to /v0/agents/{id}/followup with streaming
}

func (p *CursorProvider) GetMessages(ctx context.Context, agentID string) ([]Message, error) {
    // GET /v0/agents/{id}/messages
}
```

### 2. HTTP Client Setup

Create `backend/internal/providers/cursor/client.go`:
- Configure HTTP client with timeouts
- Handle Basic auth encoding
- Parse SSE responses for streaming
- Error handling for API errors

### 3. Request/Response Types

Create `backend/internal/providers/cursor/types.go`:
```go
type CreateAgentRequest struct {
    Prompt      string `json:"prompt"`
    Model       string `json:"model,omitempty"`
    WorkspaceID string `json:"workspace_id,omitempty"`
}

type AgentResponse struct {
    ID        string `json:"id"`
    Status    string `json:"status"`
    CreatedAt string `json:"created_at"`
}

type MessageResponse struct {
    ID        string `json:"id"`
    Role      string `json:"role"`
    Content   string `json:"content"`
    CreatedAt string `json:"created_at"`
}
```

### 4. Streaming Support

- Parse SSE events from Cursor API
- Convert to internal `StreamChunk` format
- Handle connection drops and reconnection

### 5. API Key Management

- Store encrypted in `provider_keys` table
- Use existing encryption service
- Provider-specific key validation

## Files to Create/Modify

- `backend/internal/providers/cursor/provider.go` - Main provider implementation
- `backend/internal/providers/cursor/client.go` - HTTP client utilities
- `backend/internal/providers/cursor/types.go` - Request/response types
- `backend/internal/providers/cursor/streaming.go` - SSE parsing
- `backend/internal/providers/manager.go` - Register Cursor provider

## Acceptance Criteria

- [ ] CursorProvider implements AgentProvider interface
- [ ] Basic auth with API key works correctly
- [ ] Can create agents via Cursor API
- [ ] Can retrieve message history
- [ ] Can send follow-up messages with streaming
- [ ] SSE streaming parsed correctly
- [ ] API errors handled gracefully
- [ ] API key stored encrypted in database
- [ ] Provider registered with provider manager
- [ ] Unit tests with mocked HTTP responses

## Integration Points

- **Provides**: Cursor agent creation/messaging capability
- **Consumes**: AgentProvider interface from prism-provider task
- **Conflicts**: None - isolated external API client
