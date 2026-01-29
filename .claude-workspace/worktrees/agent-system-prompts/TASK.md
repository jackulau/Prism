---
id: agent-system-prompts
name: Agent System Prompts for Standard and Orchestrator Agents
wave: 1
priority: 1
dependencies: []
estimated_hours: 3
tags:
- backend
- agent
- prompts
---

## Objective

Create comprehensive system prompts for the standard coding agent and orchestrator multi-agent coordinator.

## Context

The codebase already has role-based prompts in `/backend/internal/agent/orchestrator.go` for roles like Planner, Coder, Reviewer, etc. This task adds two new specialized prompts:
1. **AGENT_SYSTEM_PROMPT** - For standard coding agents that execute tasks with tools
2. **ORCHESTRATOR_AGENT_SYSTEM_PROMPT** - For multi-agent coordinators that spawn and manage sub-agents

## Implementation

1. Create `/backend/internal/agent/prompts.go` with:
   - `AGENT_SYSTEM_PROMPT` - Standard coding agent prompt covering:
     - Code execution capabilities
     - Tool usage guidelines
     - Security considerations
     - Best practices for file modifications
     - Error handling guidance
   - `ORCHESTRATOR_AGENT_SYSTEM_PROMPT` - Coordinator prompt covering:
     - Sub-agent spawning capabilities
     - Task decomposition strategies
     - Result aggregation patterns
     - Coordination workflows

2. Export prompts as constants for use in agent configuration

3. Add unit tests in `/backend/internal/agent/prompts_test.go`:
   - Verify prompts are non-empty
   - Verify key content elements exist
   - Test prompt concatenation with context

## Acceptance Criteria

- [ ] `AGENT_SYSTEM_PROMPT` constant defined with comprehensive coding agent instructions
- [ ] `ORCHESTRATOR_AGENT_SYSTEM_PROMPT` constant defined with coordination instructions
- [ ] Prompts follow existing codebase patterns (see `orchestrator.go` role prompts)
- [ ] Unit tests pass
- [ ] No breaking changes to existing prompt system

## Files to Create/Modify

- `backend/internal/agent/prompts.go` - Create new file with prompt constants
- `backend/internal/agent/prompts_test.go` - Create tests for prompts

## Integration Points

- **Provides**: System prompt constants for agent configuration
- **Consumes**: None (foundational)
- **Conflicts**: None - new file, does not modify existing code
