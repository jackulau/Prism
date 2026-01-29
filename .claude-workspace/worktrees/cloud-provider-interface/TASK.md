---
id: cloud-provider-interface
name: CloudProvider Interface + Prism Provider
wave: 1
priority: 1
dependencies: []
estimated_hours: 4
tags:
- backend
- core
- provider
---

## Objective

Create a CloudProvider interface abstraction and implement the Prism Provider for cloud-hosted AI services.

## Context

The codebase already has a well-designed Provider interface in `backend/internal/llm/provider.go` for LLM services. We need to create a similar CloudProvider interface that extends this pattern to support cloud-hosted AI agent services like Prism (Claude-based cloud provider). This will enable the system to route requests to either local LLM providers or cloud-hosted agent services.

## Implementation

### 1. Create CloudProvider Interface

**File**: `backend/internal/cloud/provider.go`

```go
package cloud

import (
    "context"
)

type CloudProvider interface {
    Name() string
    Connect(ctx context.Context) error
    Disconnect() error
    CreateSession(ctx context.Context, opts *SessionOptions) (*Session, error)
    SendMessage(ctx context.Context, sessionID string, message *Message) (<-chan StreamChunk, error)
    GetSession(ctx context.Context, sessionID string) (*Session, error)
    ListSessions(ctx context.Context) ([]*Session, error)
    DeleteSession(ctx context.Context, sessionID string) error
    IsConnected() bool
    ValidateCredentials(ctx context.Context) error
}

type Session struct {
    ID        string
    Provider  string
    Status    SessionStatus
    CreatedAt time.Time
    UpdatedAt time.Time
    Metadata  map[string]interface{}
}

type SessionOptions struct {
    Model       string
    SystemPrompt string
    Temperature float64
    MaxTokens   int
}

type StreamChunk struct {
    Delta        string
    ToolCalls    []ToolCall
    FinishReason string
    Error        error
}
```

### 2. Create CloudManager

**File**: `backend/internal/cloud/manager.go`

```go
package cloud

type Manager struct {
    providers map[string]CloudProvider
    mu        sync.RWMutex
}

func NewManager() *Manager
func (m *Manager) RegisterProvider(provider CloudProvider) error
func (m *Manager) GetProvider(name string) (CloudProvider, error)
func (m *Manager) ListProviders() []string
```

### 3. Implement Prism Provider

**File**: `backend/internal/cloud/prism/client.go`

```go
package prism

type Client struct {
    apiKey    string
    baseURL   string
    client    *http.Client
    sessions  map[string]*cloud.Session
    mu        sync.RWMutex
}

func NewClient(apiKey string) *Client
func (c *Client) Name() string  // Returns "prism"
func (c *Client) Connect(ctx context.Context) error
func (c *Client) CreateSession(ctx context.Context, opts *cloud.SessionOptions) (*cloud.Session, error)
func (c *Client) SendMessage(ctx context.Context, sessionID string, message *cloud.Message) (<-chan cloud.StreamChunk, error)
// ... implement all CloudProvider methods
```

### 4. Integrate with Main Application

**File**: `backend/cmd/server/main.go`

Add CloudManager initialization alongside LLMManager.

### 5. Add API Routes

**File**: `backend/internal/api/routes/cloud_routes.go`

```go
// Routes for cloud provider management
POST /api/v1/cloud/:provider/connect
POST /api/v1/cloud/:provider/disconnect
GET  /api/v1/cloud/:provider/status
POST /api/v1/cloud/:provider/sessions
```

## Acceptance Criteria

- [ ] CloudProvider interface defined with all necessary methods
- [ ] CloudManager can register and retrieve providers
- [ ] Prism provider implements full CloudProvider interface
- [ ] Streaming works via channel pattern (matching LLM provider pattern)
- [ ] Session management (create, get, list, delete) working
- [ ] API routes for cloud provider management
- [ ] Credentials validation endpoint
- [ ] Error handling consistent with existing patterns

## Files to Create/Modify

- `backend/internal/cloud/provider.go` - CloudProvider interface
- `backend/internal/cloud/manager.go` - Provider manager
- `backend/internal/cloud/types.go` - Shared types (Session, Message, etc.)
- `backend/internal/cloud/prism/client.go` - Prism provider implementation
- `backend/internal/api/routes/cloud_routes.go` - API routes
- `backend/internal/api/handlers/cloud.go` - Request handlers
- `backend/cmd/server/main.go` - Register CloudManager

## Integration Points

- **Provides**: CloudProvider interface for cloud AI services
- **Provides**: Prism provider for Claude-based cloud agent
- **Consumes**: Config for API keys and endpoints
- **Conflicts**: Avoid modifying `internal/llm/` (separate concern)

## Notes

- Follow existing patterns from `internal/llm/provider.go` for interface design
- Use `sync.RWMutex` for thread-safe provider management
- Implement streaming with Go channels like OpenAI client does
- Store encrypted credentials using existing EncryptionService
