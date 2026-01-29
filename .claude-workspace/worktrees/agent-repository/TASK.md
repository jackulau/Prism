---
id: agent-repository
name: Agent Repository Layer
wave: 2
priority: 2
dependencies:
- agent-db-schema
estimated_hours: 3
tags:
- backend
- repository
---

## Objective

Create the Agent repository with CRUD operations following the existing repository pattern in the codebase.

## Context

The codebase uses a repository pattern in `/backend/internal/database/repository/`. Each entity has a dedicated repository file with standard methods. This task creates `agent.go` following patterns from `conversation.go`, `todo.go`, and `workspace.go`.

## Implementation

1. **Create `agent.go`** in `/backend/internal/database/repository/`

2. **Define Agent struct**:
```go
type Agent struct {
    ID                  string
    WorkspaceID         *string  // nullable
    Status              string   // PENDING, RUNNING, COMPLETED, FAILED
    ProviderType        string   // PRISM, CURSOR, JULES
    ConversationID      *string  // nullable
    URL                 *string  // nullable
    GithubBranchName    *string  // nullable
    Name                string
    Model               *string  // nullable
    SandboxID           *string  // nullable
    IsOrchestratorAgent bool
    CreatedAt           time.Time
    UpdatedAt           time.Time
}
```

3. **Implement AgentRepository**:
```go
type AgentRepository struct {
    db *sql.DB
}

func NewAgentRepository(db *sql.DB) *AgentRepository
func (r *AgentRepository) Create(agent *Agent) (*Agent, error)
func (r *AgentRepository) GetByID(id string) (*Agent, error)
func (r *AgentRepository) ListByWorkspaceID(workspaceID string, limit, offset int) ([]*Agent, error)
func (r *AgentRepository) ListByStatus(status string, limit, offset int) ([]*Agent, error)
func (r *AgentRepository) Update(agent *Agent) error
func (r *AgentRepository) UpdateStatus(id, status string) error
func (r *AgentRepository) Delete(id string) error
func (r *AgentRepository) DeleteByWorkspaceID(workspaceID string) error
```

4. **Use uuid.New().String()** for ID generation

5. **Handle nullable fields** with `sql.NullString` for scanning

## Acceptance Criteria

- [ ] `agent.go` created in repository package
- [ ] Agent struct matches database schema
- [ ] All CRUD methods implemented
- [ ] Uses parameterized queries (? placeholders) for SQL injection prevention
- [ ] Proper error handling with wrapped errors
- [ ] Nullable fields handled with sql.Null* types
- [ ] UpdatedAt set automatically on Update operations
- [ ] UUID generation for new agents

## Files to Create/Modify

- `backend/internal/database/repository/agent.go` - Create new file

## Integration Points

- **Provides**: AgentRepository for API handlers
- **Consumes**: `agents` table from agent-db-schema task
- **Conflicts**: None - new file
