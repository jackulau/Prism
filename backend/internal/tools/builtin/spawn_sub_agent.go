package builtin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/agent"
	"github.com/jacklau/prism/internal/llm"
)

// SpawnSubAgentTool allows an orchestrator to spawn sub-agents for specific tasks
type SpawnSubAgentTool struct {
	llmManager *llm.Manager
	workflow   *agent.OrchestratorWorkflow
}

// NewSpawnSubAgentTool creates a new spawn sub-agent tool
func NewSpawnSubAgentTool(llmManager *llm.Manager) *SpawnSubAgentTool {
	return &SpawnSubAgentTool{
		llmManager: llmManager,
	}
}

// SetWorkflow sets the orchestrator workflow for this tool
// This allows the tool to spawn sub-agents within the workflow context
func (t *SpawnSubAgentTool) SetWorkflow(workflow *agent.OrchestratorWorkflow) {
	t.workflow = workflow
}

func (t *SpawnSubAgentTool) Name() string {
	return "spawn_sub_agent"
}

func (t *SpawnSubAgentTool) Description() string {
	return `Spawn a specialized sub-agent to handle a specific subtask. Use this when you need to delegate work to a specialized agent.

Available roles:
- general: General-purpose agent for various tasks
- planner: Task breakdown and planning specialist
- coder: Code writing and implementation specialist
- reviewer: Code review and quality analysis specialist
- researcher: Information gathering and research specialist
- writer: Documentation and content writing specialist
- analyst: Data analysis and insight specialist
- debugger: Bug identification and fixing specialist
- tester: Test design and verification specialist
- synthesizer: Result combination and synthesis specialist

The sub-agent will execute the task and return its results.`
}

func (t *SpawnSubAgentTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"task_prompt": {
				Type:        "string",
				Description: "The task prompt describing what the sub-agent should do",
			},
			"role": {
				Type:        "string",
				Description: "The role/specialization of the sub-agent",
				Enum: []string{
					"general", "planner", "coder", "reviewer",
					"researcher", "writer", "analyst", "debugger",
					"tester", "synthesizer",
				},
			},
			"context": {
				Type:        "string",
				Description: "Optional additional context to provide to the sub-agent",
			},
			"provider": {
				Type:        "string",
				Description: "Optional LLM provider override (e.g., 'openai', 'anthropic'). Uses orchestrator's provider if not specified.",
			},
			"model": {
				Type:        "string",
				Description: "Optional model override. Uses orchestrator's model if not specified.",
			},
		},
		Required: []string{"task_prompt", "role"},
	}
}

func (t *SpawnSubAgentTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	taskPrompt, ok := params["task_prompt"].(string)
	if !ok || taskPrompt == "" {
		return nil, fmt.Errorf("task_prompt parameter is required")
	}

	roleStr, ok := params["role"].(string)
	if !ok || roleStr == "" {
		return nil, fmt.Errorf("role parameter is required")
	}

	// Validate role
	role := agent.AgentRole(roleStr)
	if !isValidRole(role) {
		return nil, fmt.Errorf("invalid role: %s", roleStr)
	}

	// Get optional parameters
	contextStr := ""
	if c, ok := params["context"].(string); ok {
		contextStr = c
	}

	provider := ""
	if p, ok := params["provider"].(string); ok {
		provider = p
	}

	model := ""
	if m, ok := params["model"].(string); ok {
		model = m
	}

	// If workflow is set, use it to spawn the sub-agent
	if t.workflow != nil {
		return t.spawnWithWorkflow(ctx, taskPrompt, role, contextStr, provider, model)
	}

	// Otherwise, spawn a standalone sub-agent
	return t.spawnStandalone(ctx, taskPrompt, role, contextStr, provider, model)
}

func (t *SpawnSubAgentTool) spawnWithWorkflow(ctx context.Context, taskPrompt string, role agent.AgentRole, context, provider, model string) (interface{}, error) {
	// Build agent config from workflow config with overrides
	agentConfig := agent.AgentConfig{
		ID:           uuid.New().String(),
		Name:         string(role) + "-agent",
		Provider:     provider,
		Model:        model,
		SystemPrompt: getRoleSystemPrompt(role),
	}

	// Create task
	task := agent.NewTask(taskPrompt)
	if context != "" {
		task = agent.NewTask(taskPrompt, agent.WithContext(context))
	}

	// Spawn sub-agent through workflow
	result, err := t.workflow.SpawnSubAgent(agentConfig, task)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"success":      result.Success,
		"agent_id":     result.AgentID,
		"role":         string(role),
		"output":       result.Output,
		"error":        result.Error,
		"duration_ms":  result.Duration.Milliseconds(),
		"tool_results": result.ToolResults,
	}, nil
}

func (t *SpawnSubAgentTool) spawnStandalone(ctx context.Context, taskPrompt string, role agent.AgentRole, contextStr, provider, model string) (interface{}, error) {
	if t.llmManager == nil {
		return nil, fmt.Errorf("LLM manager not configured")
	}

	// Use defaults if not specified
	if provider == "" {
		provider = "anthropic" // Default provider
	}
	if model == "" {
		model = "claude-sonnet-4-20250514" // Default model
	}

	agentConfig := agent.AgentConfig{
		ID:           uuid.New().String(),
		Name:         string(role) + "-agent",
		Provider:     provider,
		Model:        model,
		SystemPrompt: getRoleSystemPrompt(role),
	}

	// Create the agent
	subAgent := agent.NewAgent(agentConfig, t.llmManager)

	// Create task
	task := agent.NewTask(taskPrompt)
	if contextStr != "" {
		task = agent.NewTask(taskPrompt, agent.WithContext(contextStr))
	}

	// Start the agent
	startTime := time.Now()
	if err := subAgent.Start(ctx, task); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	// Wait for result with timeout
	select {
	case result := <-subAgent.Results():
		return map[string]interface{}{
			"success":      result.Success,
			"agent_id":     result.AgentID,
			"role":         string(role),
			"output":       result.Output,
			"error":        result.Error,
			"duration_ms":  result.Duration.Milliseconds(),
			"tool_results": result.ToolResults,
		}, nil

	case <-ctx.Done():
		subAgent.Stop()
		return map[string]interface{}{
			"success":     false,
			"agent_id":    agentConfig.ID,
			"role":        string(role),
			"error":       "context cancelled or timeout",
			"duration_ms": time.Since(startTime).Milliseconds(),
		}, nil
	}
}

func (t *SpawnSubAgentTool) RequiresConfirmation() bool {
	return false // Sub-agent spawning is controlled by the orchestrator
}

// isValidRole checks if the given role is valid
func isValidRole(role agent.AgentRole) bool {
	validRoles := map[agent.AgentRole]bool{
		agent.RoleGeneral:     true,
		agent.RolePlanner:     true,
		agent.RoleCoder:       true,
		agent.RoleReviewer:    true,
		agent.RoleResearcher:  true,
		agent.RoleWriter:      true,
		agent.RoleAnalyst:     true,
		agent.RoleDebugger:    true,
		agent.RoleTester:      true,
		agent.RoleSynthesizer: true,
	}
	return validRoles[role]
}

// getRoleSystemPrompt returns the default system prompt for a given role
func getRoleSystemPrompt(role agent.AgentRole) string {
	prompts := map[agent.AgentRole]string{
		agent.RoleGeneral: "You are a helpful AI assistant. Complete the task to the best of your ability.",

		agent.RolePlanner: `You are a planning specialist. Your role is to:
- Break down complex tasks into actionable steps
- Identify dependencies between tasks
- Create clear, structured plans
- Prioritize tasks effectively
Focus on creating practical, implementable plans.`,

		agent.RoleCoder: `You are an expert software developer. Your role is to:
- Write clean, efficient, well-documented code
- Follow best practices and design patterns
- Consider edge cases and error handling
- Write code that is maintainable and testable
Focus on producing high-quality, working code.`,

		agent.RoleReviewer: `You are a code review specialist. Your role is to:
- Identify bugs, security issues, and code smells
- Suggest improvements for readability and performance
- Verify adherence to best practices
- Provide constructive, actionable feedback
Be thorough but constructive in your reviews.`,

		agent.RoleResearcher: `You are a research specialist. Your role is to:
- Gather and synthesize information on topics
- Identify relevant sources and references
- Provide comprehensive analysis
- Present findings clearly and objectively
Focus on accuracy and thoroughness.`,

		agent.RoleWriter: `You are a technical writer. Your role is to:
- Create clear, well-structured documentation
- Explain complex concepts simply
- Write for the target audience
- Ensure consistency and accuracy
Focus on clarity and readability.`,

		agent.RoleAnalyst: `You are an analytical specialist. Your role is to:
- Analyze data and identify patterns
- Evaluate options and trade-offs
- Provide data-driven recommendations
- Present analysis clearly
Focus on insights and actionable conclusions.`,

		agent.RoleDebugger: `You are a debugging specialist. Your role is to:
- Identify root causes of issues
- Trace execution flow and state
- Propose targeted fixes
- Verify fixes don't introduce new issues
Be systematic and thorough in debugging.`,

		agent.RoleTester: `You are a testing specialist. Your role is to:
- Design comprehensive test cases
- Identify edge cases and failure modes
- Write clear test specifications
- Verify functionality meets requirements
Focus on coverage and reliability.`,

		agent.RoleSynthesizer: `You are a synthesis specialist. Your role is to:
- Combine multiple inputs into coherent output
- Identify common themes and key points
- Resolve conflicts between different perspectives
- Create a unified, comprehensive response
Focus on creating a cohesive final result.`,
	}

	if prompt, ok := prompts[role]; ok {
		return prompt
	}
	return prompts[agent.RoleGeneral]
}
