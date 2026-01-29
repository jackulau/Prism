---
id: workflow-step-functions
name: Modular Workflow Step Functions for Agent Lifecycle
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

Implement modular workflow step functions that handle the complete agent lifecycle: loading, validation, sandbox preparation, message history, branch management, response saving, commits, and cleanup.

## Context

The codebase has foundational components:
- Agent management in `/backend/internal/agent/` (agent.go, manager.go, pool.go)
- Sandbox service in `/backend/internal/sandbox/sandbox.go`
- Message repository in `/backend/internal/database/repository/`
- JWT/auth in `/backend/internal/security/jwt.go`

These step functions extract and modularize common workflow patterns for reusable agent execution.

## Implementation

1. Create `/backend/internal/agent/workflow_steps.go` with these functions:

   ```go
   // loadAgent retrieves agent configuration by ID from storage
   func loadAgent(agentId string) (*Agent, error)
   
   // validateAndGetToken validates repo access and returns GitHub token
   func validateAndGetToken(repoName string) (string, error)
   
   // prepareSandbox creates/configures sandbox for agent execution
   func prepareSandbox(agentId string, repo string, token string) (*sandbox.Service, error)
   
   // loadPreviousMessages retrieves conversation history for agent
   func loadPreviousMessages(agentId string) ([]llm.Message, error)
   
   // createBranchIfNeeded creates a feature branch for agent work
   func createBranchIfNeeded(agent *Agent, sandbox *sandbox.Service) (string, error)
   
   // saveAgentResponse persists agent response and usage stats
   func saveAgentResponse(agentId string, text string, usage *llm.Usage) error
   
   // commitChangesIfNeeded commits sandbox changes with generated message
   func commitChangesIfNeeded(sandbox *sandbox.Service, agent *Agent, prompt string) error
   
   // cleanupSandbox releases sandbox resources
   func cleanupSandbox(sandbox *sandbox.Service) error
   ```

2. Each function should:
   - Have clear input/output contracts
   - Handle errors gracefully with proper error types
   - Log operations for debugging
   - Be independently testable

3. Add tests in `/backend/internal/agent/workflow_steps_test.go`:
   - Unit tests with mocked dependencies
   - Integration tests for sandbox/git operations

## Acceptance Criteria

- [ ] All 8 workflow step functions implemented
- [ ] Functions use existing infrastructure (sandbox, repositories)
- [ ] Error handling with typed errors
- [ ] Unit tests with mocked dependencies
- [ ] Integration with existing agent lifecycle
- [ ] Branch naming follows `prism/{agent-id}-{sanitized-prompt}` format
- [ ] Commit messages generated from prompt (truncated to 50 chars)

## Files to Create/Modify

- `backend/internal/agent/workflow_steps.go` - Create with step functions
- `backend/internal/agent/workflow_steps_test.go` - Create tests
- `backend/internal/agent/errors.go` - Add new error types if needed

## Integration Points

- **Provides**: Modular workflow steps for agent execution
- **Consumes**: sandbox.Service, MessageRepository, WorkspaceRepository
- **Conflicts**: Avoid modifying existing agent.go execution loop directly
