---
id: agent-workflow-executor
name: Standard Agent Workflow Executor
wave: 1
priority: 1
dependencies: []
estimated_hours: 8
tags:
- backend
- agents
- workflow
---

## Objective

Implement the standard agent workflow executor that handles the complete lifecycle of coding agent execution, from loading agent configuration to committing changes and cleanup.

## Context

This is the core workflow that all coding agents (Prism native and potentially delegated to external providers) follow. It orchestrates the 10-step process of executing an agent task within a sandboxed environment, handling GitHub integration, LLM interactions, and proper resource cleanup.

## Workflow Steps (10-Step Process)

1. **Load agent from DB** - Retrieve agent configuration and state
2. **Validate and get GitHub token** - Ensure user has valid GitHub credentials
3. **Create sandbox environment** - Set up isolated workspace
4. **Load conversation history** - Retrieve previous messages for context
5. **Create branch if first run** - Initialize feature branch for new agents
6. **Run LLM with tools** - Execute the agent loop with tool calling
7. **Save response and track tokens** - Persist results and usage metrics
8. **Commit changes if any** - Auto-commit file modifications to branch
9. **Cleanup sandbox** - Release resources and temporary files
10. **Mark complete** - Update agent status in database

## Implementation

### 1. Create Workflow Executor

Create `backend/internal/workflow/executor.go`:
```go
type WorkflowExecutor struct {
    agentManager  *agent.Manager
    llmManager    *llm.Manager
    sandbox       *sandbox.Service
    githubClient  *github.Client
    repos         *WorkflowRepositories
}

type WorkflowRepositories struct {
    Agent        *repository.AgentRepository
    Conversation *repository.ConversationRepository
    Message      *repository.MessageRepository
    GitHub       *repository.GitHubRepository
    TokenUsage   *repository.TokenUsageRepository
}

func (e *WorkflowExecutor) Execute(ctx context.Context, agentID string, userID string) error {
    // Orchestrate 10-step workflow
}
```

### 2. Step 1-2: Load and Validate

Create `backend/internal/workflow/steps/load.go`:
```go
func (e *WorkflowExecutor) loadAgent(ctx context.Context, agentID string) (*Agent, error)
func (e *WorkflowExecutor) validateGitHubToken(ctx context.Context, userID string) (string, error)
```

### 3. Step 3-4: Sandbox and History

Create `backend/internal/workflow/steps/setup.go`:
```go
func (e *WorkflowExecutor) createSandbox(ctx context.Context, userID string, repoURL string) (*SandboxContext, error)
func (e *WorkflowExecutor) loadConversationHistory(ctx context.Context, conversationID string) ([]llm.Message, error)
```

### 4. Step 5: Branch Creation

Create `backend/internal/workflow/steps/git.go`:
```go
func (e *WorkflowExecutor) createBranchIfNeeded(ctx context.Context, sandboxCtx *SandboxContext, agentID string) error
func (e *WorkflowExecutor) commitChanges(ctx context.Context, sandboxCtx *SandboxContext, message string) error
```

### 5. Step 6: LLM Execution Loop

Create `backend/internal/workflow/steps/llm.go`:
```go
func (e *WorkflowExecutor) runLLMLoop(ctx context.Context, config LLMLoopConfig) (*LLMResult, error)

type LLMLoopConfig struct {
    Agent         *Agent
    Messages      []llm.Message
    Tools         []llm.ToolDefinition
    MaxIterations int
    SandboxCtx    *SandboxContext
}

type LLMResult struct {
    Output     string
    ToolCalls  []ToolCallResult
    TokenUsage *llm.Usage
    Iterations int
}
```

### 6. Step 7-8: Persistence and Commit

Create `backend/internal/workflow/steps/persist.go`:
```go
func (e *WorkflowExecutor) saveResponse(ctx context.Context, agentID string, result *LLMResult) error
func (e *WorkflowExecutor) trackTokens(ctx context.Context, userID, agentID string, usage *llm.Usage) error
```

### 7. Step 9-10: Cleanup and Complete

Create `backend/internal/workflow/steps/cleanup.go`:
```go
func (e *WorkflowExecutor) cleanupSandbox(ctx context.Context, sandboxCtx *SandboxContext) error
func (e *WorkflowExecutor) markComplete(ctx context.Context, agentID string, status string, error string) error
```

### 8. Event Streaming

Create `backend/internal/workflow/events.go`:
```go
type WorkflowEvent struct {
    AgentID   string
    Step      string
    Status    string
    Data      map[string]interface{}
    Timestamp time.Time
}

type WorkflowEventHandler func(event WorkflowEvent)
```

### 9. Error Recovery

Create `backend/internal/workflow/recovery.go`:
```go
func (e *WorkflowExecutor) handleError(ctx context.Context, agentID string, step string, err error) error
func (e *WorkflowExecutor) rollback(ctx context.Context, agentID string, step string) error
```

## Files to Create/Modify

- `backend/internal/workflow/executor.go` - Main workflow orchestrator
- `backend/internal/workflow/types.go` - Workflow types and interfaces
- `backend/internal/workflow/events.go` - Event streaming
- `backend/internal/workflow/recovery.go` - Error handling and recovery
- `backend/internal/workflow/steps/load.go` - Steps 1-2
- `backend/internal/workflow/steps/setup.go` - Steps 3-4
- `backend/internal/workflow/steps/git.go` - Step 5, 8
- `backend/internal/workflow/steps/llm.go` - Step 6
- `backend/internal/workflow/steps/persist.go` - Step 7
- `backend/internal/workflow/steps/cleanup.go` - Steps 9-10
- `backend/internal/database/repository/agent_execution.go` - Agent execution tracking
- `backend/internal/database/repository/token_usage.go` - Token usage tracking

## Database Schema Additions

```sql
-- Agent execution tracking
CREATE TABLE IF NOT EXISTS agent_executions (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    conversation_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    current_step TEXT,
    started_at DATETIME,
    completed_at DATETIME,
    error TEXT,
    branch_name TEXT,
    commit_sha TEXT,
    iterations INTEGER DEFAULT 0,
    FOREIGN KEY (agent_id) REFERENCES agents(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Token usage tracking
CREATE TABLE IF NOT EXISTS token_usage (
    id TEXT PRIMARY KEY,
    execution_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    prompt_tokens INTEGER DEFAULT 0,
    completion_tokens INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    cost_usd REAL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (execution_id) REFERENCES agent_executions(id)
);
```

## Acceptance Criteria

- [ ] WorkflowExecutor implements all 10 steps
- [ ] Agent loaded from database correctly
- [ ] GitHub token validated before operations
- [ ] Sandbox created with proper isolation
- [ ] Conversation history loaded into context
- [ ] Branch created for new agents
- [ ] LLM loop executes with tool support
- [ ] Token usage tracked and persisted
- [ ] Changes committed to branch
- [ ] Sandbox cleaned up after execution
- [ ] Agent status updated on completion
- [ ] Events emitted for each step
- [ ] Error recovery handles failures gracefully
- [ ] Unit tests for each workflow step

## Integration Points

- **Provides**: Standard agent execution workflow
- **Consumes**: `agent.Manager`, `llm.Manager`, `sandbox.Service`, GitHub client, repositories
- **Conflicts**: May need to integrate with existing agent execution in `agent/manager.go`
