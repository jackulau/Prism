---
id: agent-workflows
name: Agent Workflows System
wave: 1
priority: 1
dependencies: []
estimated_hours: 6
tags:
- backend
- agent
- workflow
---

## Objective

Implement a comprehensive agent workflow system that enables structured, multi-step AI agent execution with state management, checkpointing, and resumability.

## Context

The codebase already has an agent system in `backend/internal/agent/` with:
- `agent.go` - Basic agent with event-driven architecture
- `orchestrator.go` - Multi-agent swarm orchestration with strategies (parallel, pipeline, debate, etc.)
- `manager.go` - Agent pool and task queue management

We need to extend this with a proper workflow system that supports:
- Declarative workflow definitions
- Step-by-step execution with state persistence
- Workflow resumption after interruption
- Conditional branching and loops
- Workflow templates for common patterns

## Implementation

### 1. Create Workflow Types

**File**: `backend/internal/agent/workflow/types.go`

```go
package workflow

type WorkflowStatus string
const (
    StatusPending    WorkflowStatus = "pending"
    StatusRunning    WorkflowStatus = "running"
    StatusPaused     WorkflowStatus = "paused"
    StatusCompleted  WorkflowStatus = "completed"
    StatusFailed     WorkflowStatus = "failed"
    StatusCancelled  WorkflowStatus = "cancelled"
)

type Workflow struct {
    ID          string
    Name        string
    Description string
    Steps       []Step
    Status      WorkflowStatus
    CurrentStep int
    State       map[string]interface{}
    CreatedAt   time.Time
    UpdatedAt   time.Time
    CompletedAt *time.Time
    Error       string
}

type Step struct {
    ID          string
    Name        string
    Type        StepType
    Config      StepConfig
    Condition   *Condition  // Optional: skip if condition false
    OnSuccess   string      // Next step ID on success
    OnFailure   string      // Next step ID on failure
    Timeout     time.Duration
    RetryPolicy *RetryPolicy
}

type StepType string
const (
    StepTypeAgent     StepType = "agent"      // Run agent with prompt
    StepTypeTool      StepType = "tool"       // Execute specific tool
    StepTypeCondition StepType = "condition"  // Evaluate condition
    StepTypeParallel  StepType = "parallel"   // Run multiple steps in parallel
    StepTypeWait      StepType = "wait"       // Wait for external input
    StepTypeTransform StepType = "transform"  // Transform data
)

type StepConfig struct {
    AgentConfig  *AgentStepConfig
    ToolConfig   *ToolStepConfig
    ParallelConfig *ParallelStepConfig
    // ... other configs
}
```

### 2. Create Workflow Engine

**File**: `backend/internal/agent/workflow/engine.go`

```go
package workflow

type Engine struct {
    repository  WorkflowRepository
    agentMgr    *agent.Manager
    toolRegistry *tools.Registry
    events      chan *WorkflowEvent
    mu          sync.RWMutex
}

func NewEngine(repo WorkflowRepository, agentMgr *agent.Manager) *Engine
func (e *Engine) CreateWorkflow(def *WorkflowDefinition) (*Workflow, error)
func (e *Engine) StartWorkflow(ctx context.Context, workflowID string) error
func (e *Engine) PauseWorkflow(ctx context.Context, workflowID string) error
func (e *Engine) ResumeWorkflow(ctx context.Context, workflowID string) error
func (e *Engine) CancelWorkflow(ctx context.Context, workflowID string) error
func (e *Engine) GetWorkflow(workflowID string) (*Workflow, error)
func (e *Engine) ListWorkflows(filter *WorkflowFilter) ([]*Workflow, error)
func (e *Engine) Events() <-chan *WorkflowEvent
```

### 3. Create Step Executors

**File**: `backend/internal/agent/workflow/executors.go`

```go
package workflow

type StepExecutor interface {
    Execute(ctx context.Context, step *Step, state map[string]interface{}) (*StepResult, error)
}

type AgentExecutor struct { /* ... */ }
type ToolExecutor struct { /* ... */ }
type ParallelExecutor struct { /* ... */ }
type ConditionExecutor struct { /* ... */ }
```

### 4. Create Workflow Repository

**File**: `backend/internal/database/repository/workflow.go`

```go
type WorkflowRepository struct {
    db *sql.DB
}

func (r *WorkflowRepository) Create(workflow *Workflow) error
func (r *WorkflowRepository) GetByID(id string) (*Workflow, error)
func (r *WorkflowRepository) Update(workflow *Workflow) error
func (r *WorkflowRepository) UpdateState(id string, state map[string]interface{}) error
func (r *WorkflowRepository) List(filter *WorkflowFilter) ([]*Workflow, error)
func (r *WorkflowRepository) Delete(id string) error
```

### 5. Add Database Migration

**File**: `backend/internal/database/migrations/005_workflows.sql`

```sql
CREATE TABLE workflows (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    definition TEXT NOT NULL,  -- JSON
    status TEXT NOT NULL DEFAULT 'pending',
    current_step INTEGER DEFAULT 0,
    state TEXT,  -- JSON
    error TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE workflow_steps (
    id TEXT PRIMARY KEY,
    workflow_id TEXT NOT NULL,
    step_index INTEGER NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    config TEXT NOT NULL,  -- JSON
    status TEXT DEFAULT 'pending',
    result TEXT,  -- JSON
    started_at DATETIME,
    completed_at DATETIME,
    error TEXT,
    FOREIGN KEY (workflow_id) REFERENCES workflows(id)
);
```

### 6. Create Workflow Templates

**File**: `backend/internal/agent/workflow/templates.go`

```go
package workflow

// Pre-defined workflow templates
var CodeReviewWorkflow = &WorkflowDefinition{
    Name: "Code Review",
    Steps: []Step{
        {Name: "analyze", Type: StepTypeAgent, Config: {...}},
        {Name: "review", Type: StepTypeAgent, Config: {...}},
        {Name: "summarize", Type: StepTypeAgent, Config: {...}},
    },
}

var DebugWorkflow = &WorkflowDefinition{...}
var RefactorWorkflow = &WorkflowDefinition{...}
```

### 7. Add API Routes

**File**: `backend/internal/api/routes/workflow_routes.go`

```go
// Workflow management routes
POST   /api/v1/workflows           - Create workflow
GET    /api/v1/workflows           - List workflows
GET    /api/v1/workflows/:id       - Get workflow details
POST   /api/v1/workflows/:id/start - Start workflow
POST   /api/v1/workflows/:id/pause - Pause workflow
POST   /api/v1/workflows/:id/resume - Resume workflow
DELETE /api/v1/workflows/:id       - Cancel/delete workflow
GET    /api/v1/workflows/templates - List available templates
```

## Acceptance Criteria

- [ ] Workflow types defined (Workflow, Step, StepType, etc.)
- [ ] Workflow engine can create, start, pause, resume workflows
- [ ] Step executors for agent, tool, parallel, and condition steps
- [ ] State persistence in database with JSON serialization
- [ ] Workflow resumption after server restart
- [ ] Event streaming for workflow progress
- [ ] Workflow templates for common patterns
- [ ] API routes for workflow management
- [ ] WebSocket events for real-time workflow updates

## Files to Create/Modify

- `backend/internal/agent/workflow/types.go` - Workflow types
- `backend/internal/agent/workflow/engine.go` - Workflow engine
- `backend/internal/agent/workflow/executors.go` - Step executors
- `backend/internal/agent/workflow/templates.go` - Workflow templates
- `backend/internal/database/repository/workflow.go` - Repository
- `backend/internal/database/migrations/005_workflows.sql` - Migration
- `backend/internal/api/routes/workflow_routes.go` - API routes
- `backend/internal/api/handlers/workflow.go` - Request handlers

## Integration Points

- **Provides**: Workflow engine for structured agent execution
- **Provides**: Workflow templates for common patterns
- **Consumes**: Agent manager for agent step execution
- **Consumes**: Tool registry for tool step execution
- **Consumes**: Database for state persistence
- **Conflicts**: Avoid modifying core agent files (agent.go, orchestrator.go)
