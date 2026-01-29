---
id: vercel-sandbox
name: Vercel Sandbox Integration
wave: 1
priority: 1
dependencies: []
estimated_hours: 5
tags:
- backend
- frontend
- sandbox
- vercel
---

## Objective

Integrate Vercel's sandbox/preview infrastructure for running and previewing code in isolated environments, replacing or augmenting the current Docker-based sandbox.

## Context

The codebase has an existing sandbox system in `backend/internal/sandbox/sandbox.go` with:
- Docker-based code execution
- Build status tracking (pending, running, success, failed)
- Working directory management
- Output streaming for build logs
- File tree management

The frontend has `BrowserPreview.tsx` which renders previews in an iframe.

We need to integrate Vercel's sandbox infrastructure to provide:
- Faster cold starts than Docker
- Better isolation and security
- Built-in framework support (Next.js, React, etc.)
- Persistent preview URLs

## Implementation

### 1. Create Vercel Sandbox Provider

**File**: `backend/internal/sandbox/vercel/provider.go`

```go
package vercel

import (
    "context"
    "github.com/your-project/internal/sandbox"
)

type Provider struct {
    apiToken    string
    teamID      string
    client      *http.Client
    deployments map[string]*Deployment
    mu          sync.RWMutex
}

func NewProvider(apiToken, teamID string) *Provider

// Implement sandbox.Provider interface
func (p *Provider) Name() string
func (p *Provider) CreateSandbox(ctx context.Context, opts *sandbox.CreateOptions) (*sandbox.Sandbox, error)
func (p *Provider) DeploySandbox(ctx context.Context, sandboxID string, files map[string][]byte) (*sandbox.DeployResult, error)
func (p *Provider) GetLogs(ctx context.Context, sandboxID string) (<-chan sandbox.LogEntry, error)
func (p *Provider) DeleteSandbox(ctx context.Context, sandboxID string) error
func (p *Provider) GetPreviewURL(sandboxID string) string
```

### 2. Define Sandbox Provider Interface

**File**: `backend/internal/sandbox/provider.go`

```go
package sandbox

type Provider interface {
    Name() string
    CreateSandbox(ctx context.Context, opts *CreateOptions) (*Sandbox, error)
    DeploySandbox(ctx context.Context, sandboxID string, files map[string][]byte) (*DeployResult, error)
    GetLogs(ctx context.Context, sandboxID string) (<-chan LogEntry, error)
    DeleteSandbox(ctx context.Context, sandboxID string) error
    GetPreviewURL(sandboxID string) string
}

type CreateOptions struct {
    Framework    Framework  // nextjs, react, vue, etc.
    NodeVersion  string
    BuildCommand string
    OutputDir    string
}

type Framework string
const (
    FrameworkNextJS  Framework = "nextjs"
    FrameworkReact   Framework = "react"
    FrameworkVue     Framework = "vue"
    FrameworkVite    Framework = "vite"
    FrameworkStatic  Framework = "static"
)

type Sandbox struct {
    ID         string
    Provider   string
    Status     SandboxStatus
    PreviewURL string
    CreatedAt  time.Time
}

type DeployResult struct {
    DeploymentID string
    PreviewURL   string
    Status       DeploymentStatus
    BuildOutput  string
}
```

### 3. Create Sandbox Manager

**File**: `backend/internal/sandbox/manager.go`

```go
package sandbox

type Manager struct {
    providers map[string]Provider
    default   string
    mu        sync.RWMutex
}

func NewManager() *Manager
func (m *Manager) RegisterProvider(provider Provider)
func (m *Manager) SetDefault(name string)
func (m *Manager) GetProvider(name string) (Provider, error)
func (m *Manager) CreateSandbox(ctx context.Context, provider string, opts *CreateOptions) (*Sandbox, error)
```

### 4. Implement Vercel API Client

**File**: `backend/internal/sandbox/vercel/api.go`

```go
package vercel

type APIClient struct {
    baseURL string
    token   string
    client  *http.Client
}

// Vercel API endpoints
func (c *APIClient) CreateDeployment(ctx context.Context, req *CreateDeploymentRequest) (*Deployment, error)
func (c *APIClient) GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error)
func (c *APIClient) ListDeployments(ctx context.Context, projectID string) ([]*Deployment, error)
func (c *APIClient) GetBuildLogs(ctx context.Context, deploymentID string) ([]LogEntry, error)
func (c *APIClient) DeleteDeployment(ctx context.Context, deploymentID string) error
```

### 5. Update Frontend Preview Component

**File**: `frontend/src/components/sandbox/BrowserPreview.tsx`

Add support for Vercel preview URLs and deployment status:

```tsx
interface BrowserPreviewProps {
    previewUrl: string;
    provider: 'docker' | 'vercel';
    deploymentStatus?: DeploymentStatus;
    onRefresh?: () => void;
}

// Add deployment status indicator
// Add Vercel-specific loading states
// Handle Vercel preview URL format
```

### 6. Add WebSocket Events

**File**: Update `backend/internal/api/websocket/messages.go`

Add new message types:
- `sandbox.created` - Sandbox created
- `sandbox.deploying` - Deployment in progress
- `sandbox.deployed` - Deployment complete with preview URL
- `sandbox.failed` - Deployment failed
- `sandbox.logs` - Log stream chunk

### 7. Add API Routes

**File**: `backend/internal/api/routes/sandbox_routes.go`

```go
// Sandbox management routes
POST   /api/v1/sandbox                  - Create sandbox
POST   /api/v1/sandbox/:id/deploy       - Deploy files to sandbox
GET    /api/v1/sandbox/:id              - Get sandbox status
GET    /api/v1/sandbox/:id/logs         - Stream build logs
DELETE /api/v1/sandbox/:id              - Delete sandbox
GET    /api/v1/sandbox/:id/preview      - Get preview URL
```

### 8. Add Sandbox Settings

**File**: `backend/internal/config/config.go`

```go
type VercelConfig struct {
    APIToken      string `env:"VERCEL_API_TOKEN"`
    TeamID        string `env:"VERCEL_TEAM_ID"`
    DefaultRegion string `env:"VERCEL_DEFAULT_REGION" default:"iad1"`
}
```

## Acceptance Criteria

- [ ] SandboxProvider interface defined
- [ ] Vercel provider implements full interface
- [ ] Sandbox manager can route to different providers
- [ ] Vercel deployments work with preview URLs
- [ ] Build log streaming works via WebSocket
- [ ] Framework detection (Next.js, React, etc.)
- [ ] Frontend shows Vercel deployment status
- [ ] API routes for sandbox management
- [ ] Environment configuration for Vercel credentials

## Files to Create/Modify

- `backend/internal/sandbox/provider.go` - Provider interface
- `backend/internal/sandbox/manager.go` - Sandbox manager
- `backend/internal/sandbox/vercel/provider.go` - Vercel provider
- `backend/internal/sandbox/vercel/api.go` - Vercel API client
- `backend/internal/api/routes/sandbox_routes.go` - API routes
- `backend/internal/api/handlers/sandbox.go` - Request handlers
- `frontend/src/components/sandbox/BrowserPreview.tsx` - Update preview component
- `frontend/src/store/sandboxStore.ts` - Add Vercel state

## Integration Points

- **Provides**: Vercel sandbox provider for fast previews
- **Provides**: Sandbox manager for provider abstraction
- **Consumes**: Config for Vercel API credentials
- **Consumes**: WebSocket hub for log streaming
- **Conflicts**: Avoid breaking existing Docker sandbox (keep as fallback)
