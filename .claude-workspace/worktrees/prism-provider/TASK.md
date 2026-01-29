---
id: prism-provider
name: Prism Provider - Native Agent Provider
wave: 1
priority: 1
dependencies: []
estimated_hours: 6
tags:
- backend
- providers
- agents
---

## Objective

Implement a native Prism provider that acts as a built-in agent provider with full sandbox control, internal workflow execution, token tracking, cost calculation, SSE streaming, and database persistence.

## Context

The Prism provider is distinct from LLM providers (Anthropic, OpenAI, etc.) - it's an "agent provider" that wraps the internal agent execution system. This allows users to run agents through the Prism platform itself with full control over sandbox, persistence, and token tracking. It should integrate with the existing `backend/internal/agent/` package and provide a unified interface similar to how external providers like Cursor and Jules will work.

## Implementation

### 1. Create Agent Provider Interface

Create `backend/internal/providers/provider.go`:
```go
type AgentProvider interface {
    Name() string
    CreateAgent(ctx context.Context, req CreateAgentRequest) (*Agent, error)
    GetAgent(ctx context.Context, agentID string) (*Agent, error)
    SendMessage(ctx context.Context, agentID string, message string) (<-chan StreamChunk, error)
    GetMessages(ctx context.Context, agentID string) ([]Message, error)
    StopAgent(ctx context.Context, agentID string) error
    SupportsStreaming() bool
}
```

### 2. Create Prism Provider Implementation

Create `backend/internal/providers/prism/provider.go`:
- Implement `AgentProvider` interface
- Use existing `agent.Manager` for execution
- Use `sandbox.Service` for sandbox management
- Integrate with `llm.Manager` for LLM calls
- Track tokens via `llm.Usage` in database
- Calculate costs based on model pricing table

### 3. Token Tracking & Cost Calculation

Create `backend/internal/providers/prism/cost.go`:
- Define model pricing table (per 1M tokens for input/output)
- Calculate costs from `llm.Usage` data
- Store in messages table or new `agent_costs` table

### 4. SSE Streaming Support

- Use existing `StreamChunk` pattern from `llm.Provider`
- Emit agent events via channels (already in `agent.Agent`)
- WebSocket integration through existing hub

### 5. Database Persistence

Create migrations and repositories:
- `agent_executions` table for tracking agent runs
- `agent_token_usage` table for token/cost tracking
- Extend `messages` table if needed

### 6. Provider Manager

Create `backend/internal/providers/manager.go`:
- Register providers (prism, cursor, jules)
- Route requests to appropriate provider
- Unified API for all providers

## Files to Create/Modify

- `backend/internal/providers/provider.go` - Provider interface definition
- `backend/internal/providers/manager.go` - Provider manager
- `backend/internal/providers/prism/provider.go` - Prism provider implementation
- `backend/internal/providers/prism/cost.go` - Cost calculation
- `backend/internal/providers/prism/streaming.go` - SSE streaming helpers
- `backend/internal/database/sqlite.go` - Add migration for agent tables
- `backend/internal/database/repository/agent_execution.go` - Agent execution repository

## Acceptance Criteria

- [ ] AgentProvider interface defined with all necessary methods
- [ ] Prism provider implements AgentProvider interface
- [ ] Token usage tracked per agent execution
- [ ] Cost calculated based on model pricing
- [ ] SSE streaming works through agent events
- [ ] Agent executions persisted to database
- [ ] Provider manager routes to correct provider
- [ ] All existing agent functionality continues to work
- [ ] Unit tests for provider and cost calculation

## Integration Points

- **Provides**: AgentProvider interface for all agent providers
- **Consumes**: `agent.Manager`, `llm.Manager`, `sandbox.Service`
- **Conflicts**: May need to modify `agent/manager.go` slightly for integration
