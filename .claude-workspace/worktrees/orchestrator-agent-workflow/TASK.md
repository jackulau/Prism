---
id: orchestrator-agent-workflow
name: Orchestrator Agent Workflow with Sub-Agent Spawning
wave: 2
priority: 2
dependencies:
- agent-system-prompts
- workflow-step-functions
estimated_hours: 6
tags:
- backend
- agent
- orchestrator
- workflow
---

## Objective

Implement a specialized orchestrator agent workflow that coordinates multiple sub-agents with enhanced reasoning capabilities, image extraction, and detailed logging.

## Context

The codebase has an existing orchestrator in `/backend/internal/agent/orchestrator.go` with swarm strategies (parallel, pipeline, debate, etc.). This task builds a new workflow specifically designed for:
- High reasoning effort for complex coordination decisions
- Spawning sub-agents dynamically via a `spawn_sub_agent` tool
- Extracting images from conversation history for multimodal context
- Returning detailed reasoning logs for transparency

## Implementation

1. Create `/backend/internal/agent/orchestrator_workflow.go` with:

   ```go
   // OrchestratorWorkflow handles multi-agent coordination
   type OrchestratorWorkflow struct {
       agent           *Agent
       subAgents       []*Agent
       reasoningLogs   []ReasoningLog
       imageExtractor  *ImageExtractor
   }
   
   // ReasoningLog captures orchestrator decision-making
   type ReasoningLog struct {
       Timestamp   time.Time
       Decision    string
       Reasoning   string
       SubAgentID  string
   }
   
   // NewOrchestratorWorkflow creates workflow with high reasoning config
   func NewOrchestratorWorkflow(config OrchestratorConfig) *OrchestratorWorkflow
   
   // Run executes the orchestration workflow
   func (ow *OrchestratorWorkflow) Run(ctx context.Context, task *Task) (*OrchestratorResult, error)
   
   // SpawnSubAgent creates and executes a sub-agent
   func (ow *OrchestratorWorkflow) SpawnSubAgent(config AgentConfig, task *Task) (*AgentResult, error)
   
   // ExtractImages extracts images from message history
   func (ow *OrchestratorWorkflow) ExtractImages(messages []llm.Message) []ImageContent
   
   // GetReasoningLogs returns all reasoning logs
   func (ow *OrchestratorWorkflow) GetReasoningLogs() []ReasoningLog
   ```

2. Create `spawn_sub_agent` tool in `/backend/internal/tools/builtin/spawn_sub_agent.go`:
   - Tool that orchestrator can call to spawn sub-agents
   - Parameters: task prompt, agent role, optional config
   - Returns: sub-agent result or streaming handle

3. Configure orchestrator for "high" reasoning effort:
   - Set extended thinking tokens if model supports
   - Configure longer timeouts for complex decisions
   - Enable verbose logging

4. Add tests in `/backend/internal/agent/orchestrator_workflow_test.go`:
   - Test sub-agent spawning
   - Test image extraction from history
   - Test reasoning log collection
   - Integration test for full workflow

## Acceptance Criteria

- [ ] OrchestratorWorkflow implemented with sub-agent spawning
- [ ] `spawn_sub_agent` tool registered and functional
- [ ] Image extraction from multimodal message history
- [ ] Reasoning logs captured and returnable
- [ ] "High" reasoning effort configuration applied
- [ ] Unit tests with mocked sub-agents
- [ ] Integration with existing orchestrator strategies

## Files to Create/Modify

- `backend/internal/agent/orchestrator_workflow.go` - Main workflow implementation
- `backend/internal/agent/orchestrator_workflow_test.go` - Tests
- `backend/internal/tools/builtin/spawn_sub_agent.go` - Sub-agent spawning tool
- `backend/internal/tools/builtin/spawn_sub_agent_test.go` - Tool tests
- `backend/internal/tools/builtin/init.go` - Register spawn_sub_agent tool

## Integration Points

- **Provides**: High-level orchestration workflow with spawning capability
- **Consumes**: Agent, workflow step functions, system prompts
- **Conflicts**: Avoid modifying existing orchestrator.go swarm logic
