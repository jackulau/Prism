---
id: agent-api-handlers
name: Agent API Handlers & Routes
wave: 3
priority: 3
dependencies:
- agent-repository
estimated_hours: 4
tags:
- backend
- api
---

## Objective

Create REST API handlers and routes for Agent entity CRUD operations, following the Fiber v2 patterns used in the codebase.

## Context

The codebase uses Fiber v2 web framework with handlers in `/backend/internal/api/handlers/` and route registration in `/backend/internal/api/routes/router.go`. This task creates the agent handler and registers routes.

## Implementation

### 1. Create `agent.go` handler in `/backend/internal/api/handlers/`

```go
type AgentHandler struct {
    agentRepo    *repository.AgentRepository
    workspaceRepo *repository.WorkspaceRepository
}

func NewAgentHandler(agentRepo *repository.AgentRepository, workspaceRepo *repository.WorkspaceRepository) *AgentHandler

// Handlers
func (h *AgentHandler) ListAgents(c *fiber.Ctx) error      // GET /agents
func (h *AgentHandler) CreateAgent(c *fiber.Ctx) error     // POST /agents
func (h *AgentHandler) GetAgent(c *fiber.Ctx) error        // GET /agents/:id
func (h *AgentHandler) UpdateAgent(c *fiber.Ctx) error     // PATCH /agents/:id
func (h *AgentHandler) DeleteAgent(c *fiber.Ctx) error     // DELETE /agents/:id
func (h *AgentHandler) UpdateAgentStatus(c *fiber.Ctx) error // PATCH /agents/:id/status
```

### 2. Request/Response DTOs

```go
type CreateAgentRequest struct {
    WorkspaceID      *string `json:"workspace_id"`
    Name             string  `json:"name" validate:"required"`
    ProviderType     string  `json:"provider_type" validate:"required,oneof=PRISM CURSOR JULES"`
    Model            *string `json:"model"`
    IsOrchestratorAgent bool `json:"is_orchestrator_agent"`
}

type UpdateAgentRequest struct {
    Name             *string `json:"name"`
    Status           *string `json:"status" validate:"omitempty,oneof=PENDING RUNNING COMPLETED FAILED"`
    URL              *string `json:"url"`
    GithubBranchName *string `json:"github_branch_name"`
    ConversationID   *string `json:"conversation_id"`
    SandboxID        *string `json:"sandbox_id"`
    Model            *string `json:"model"`
}

type AgentResponse struct {
    ID                  string     `json:"id"`
    WorkspaceID         *string    `json:"workspace_id"`
    Status              string     `json:"status"`
    ProviderType        string     `json:"provider_type"`
    ConversationID      *string    `json:"conversation_id"`
    URL                 *string    `json:"url"`
    GithubBranchName    *string    `json:"github_branch_name"`
    Name                string     `json:"name"`
    Model               *string    `json:"model"`
    SandboxID           *string    `json:"sandbox_id"`
    IsOrchestratorAgent bool       `json:"is_orchestrator_agent"`
    CreatedAt           time.Time  `json:"created_at"`
    UpdatedAt           time.Time  `json:"updated_at"`
}
```

### 3. Update `router.go`

```go
// In Dependencies struct
AgentRepo *repository.AgentRepository

// In Setup function
agentHandler := handlers.NewAgentHandler(deps.AgentRepo, deps.WorkspaceRepo)

// Route group
agents := v1.Group("/agents", middleware.AuthMiddleware(deps.JWTService))
agents.Get("/", agentHandler.ListAgents)
agents.Post("/", agentHandler.CreateAgent)
agents.Get("/:id", agentHandler.GetAgent)
agents.Patch("/:id", agentHandler.UpdateAgent)
agents.Delete("/:id", agentHandler.DeleteAgent)
agents.Patch("/:id/status", agentHandler.UpdateAgentStatus)
```

### 4. Update `main.go`

```go
agentRepo := repository.NewAgentRepository(db.DB)

// Add to deps
AgentRepo: agentRepo,
```

## Acceptance Criteria

- [ ] `agent.go` handler created with all CRUD methods
- [ ] Request/response DTOs defined
- [ ] Input validation for required fields
- [ ] Status enum validation (PENDING, RUNNING, COMPLETED, FAILED)
- [ ] ProviderType enum validation (PRISM, CURSOR, JULES)
- [ ] Routes registered with auth middleware
- [ ] Repository initialized in main.go
- [ ] Proper error responses (400, 404, 500)
- [ ] Query params for filtering (workspace_id, status)

## Files to Create/Modify

- `backend/internal/api/handlers/agent.go` - Create new handler
- `backend/internal/api/routes/router.go` - Add routes and dependency
- `backend/cmd/server/main.go` - Initialize repository

## Integration Points

- **Provides**: REST API endpoints for Agent management
- **Consumes**: AgentRepository from agent-repository task
- **Conflicts**: Modifies router.go and main.go - coordinate with other tasks
