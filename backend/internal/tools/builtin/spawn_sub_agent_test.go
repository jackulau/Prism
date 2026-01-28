package builtin

import (
	"context"
	"testing"

	"github.com/jacklau/prism/internal/agent"
	"github.com/jacklau/prism/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpawnSubAgentTool_Name(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)
	assert.Equal(t, "spawn_sub_agent", tool.Name())
}

func TestSpawnSubAgentTool_Description(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)
	desc := tool.Description()

	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "sub-agent")
	assert.Contains(t, desc, "general")
	assert.Contains(t, desc, "planner")
	assert.Contains(t, desc, "coder")
	assert.Contains(t, desc, "reviewer")
}

func TestSpawnSubAgentTool_Parameters(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)
	params := tool.Parameters()

	assert.Equal(t, "object", params.Type)
	require.NotNil(t, params.Properties)

	// Check required parameters
	assert.Contains(t, params.Properties, "task_prompt")
	assert.Contains(t, params.Properties, "role")

	// Check optional parameters
	assert.Contains(t, params.Properties, "context")
	assert.Contains(t, params.Properties, "provider")
	assert.Contains(t, params.Properties, "model")

	// Check required fields
	assert.Contains(t, params.Required, "task_prompt")
	assert.Contains(t, params.Required, "role")

	// Check role enum values
	roleProperty := params.Properties["role"]
	assert.Contains(t, roleProperty.Enum, "general")
	assert.Contains(t, roleProperty.Enum, "planner")
	assert.Contains(t, roleProperty.Enum, "coder")
	assert.Contains(t, roleProperty.Enum, "reviewer")
	assert.Contains(t, roleProperty.Enum, "researcher")
	assert.Contains(t, roleProperty.Enum, "writer")
	assert.Contains(t, roleProperty.Enum, "analyst")
	assert.Contains(t, roleProperty.Enum, "debugger")
	assert.Contains(t, roleProperty.Enum, "tester")
	assert.Contains(t, roleProperty.Enum, "synthesizer")
}

func TestSpawnSubAgentTool_RequiresConfirmation(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)
	assert.False(t, tool.RequiresConfirmation())
}

func TestSpawnSubAgentTool_Execute_MissingTaskPrompt(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)

	params := map[string]interface{}{
		"role": "coder",
	}

	_, err := tool.Execute(context.Background(), params)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "task_prompt")
}

func TestSpawnSubAgentTool_Execute_MissingRole(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)

	params := map[string]interface{}{
		"task_prompt": "Write some code",
	}

	_, err := tool.Execute(context.Background(), params)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "role")
}

func TestSpawnSubAgentTool_Execute_InvalidRole(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)

	params := map[string]interface{}{
		"task_prompt": "Write some code",
		"role":        "invalid_role",
	}

	_, err := tool.Execute(context.Background(), params)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

func TestSpawnSubAgentTool_Execute_NoLLMManager(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)

	params := map[string]interface{}{
		"task_prompt": "Write some code",
		"role":        "coder",
	}

	_, err := tool.Execute(context.Background(), params)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "LLM manager not configured")
}

func TestSpawnSubAgentTool_SetWorkflow(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)

	config := agent.OrchestratorConfig{
		Name:     "test-orchestrator",
		Provider: "anthropic",
		Model:    "claude-sonnet-4-20250514",
	}
	workflow := agent.NewOrchestratorWorkflow(config, nil)

	tool.SetWorkflow(workflow)

	assert.Equal(t, workflow, tool.workflow)
}

func TestIsValidRole(t *testing.T) {
	validRoles := []agent.AgentRole{
		agent.RoleGeneral,
		agent.RolePlanner,
		agent.RoleCoder,
		agent.RoleReviewer,
		agent.RoleResearcher,
		agent.RoleWriter,
		agent.RoleAnalyst,
		agent.RoleDebugger,
		agent.RoleTester,
		agent.RoleSynthesizer,
	}

	for _, role := range validRoles {
		t.Run(string(role), func(t *testing.T) {
			assert.True(t, isValidRole(role))
		})
	}

	t.Run("invalid role", func(t *testing.T) {
		assert.False(t, isValidRole(agent.AgentRole("invalid")))
	})
}

func TestGetRoleSystemPrompt(t *testing.T) {
	testCases := []struct {
		role     agent.AgentRole
		contains string
	}{
		{agent.RoleGeneral, "helpful AI assistant"},
		{agent.RolePlanner, "planning specialist"},
		{agent.RoleCoder, "software developer"},
		{agent.RoleReviewer, "code review specialist"},
		{agent.RoleResearcher, "research specialist"},
		{agent.RoleWriter, "technical writer"},
		{agent.RoleAnalyst, "analytical specialist"},
		{agent.RoleDebugger, "debugging specialist"},
		{agent.RoleTester, "testing specialist"},
		{agent.RoleSynthesizer, "synthesis specialist"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.role), func(t *testing.T) {
			prompt := getRoleSystemPrompt(tc.role)
			assert.NotEmpty(t, prompt)
			assert.Contains(t, prompt, tc.contains)
		})
	}

	t.Run("unknown role returns general prompt", func(t *testing.T) {
		prompt := getRoleSystemPrompt(agent.AgentRole("unknown"))
		assert.Contains(t, prompt, "helpful AI assistant")
	})
}

func TestSpawnSubAgentTool_Execute_EmptyTaskPrompt(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)

	params := map[string]interface{}{
		"task_prompt": "",
		"role":        "coder",
	}

	_, err := tool.Execute(context.Background(), params)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "task_prompt")
}

func TestSpawnSubAgentTool_Execute_EmptyRole(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)

	params := map[string]interface{}{
		"task_prompt": "Write code",
		"role":        "",
	}

	_, err := tool.Execute(context.Background(), params)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "role")
}

func TestSpawnSubAgentTool_Execute_WithOptionalParams(t *testing.T) {
	// This test validates parameter parsing without actually executing
	// since we don't have a mock LLM manager
	tool := NewSpawnSubAgentTool(nil)

	params := map[string]interface{}{
		"task_prompt": "Write some code",
		"role":        "coder",
		"context":     "This is additional context",
		"provider":    "openai",
		"model":       "gpt-4",
	}

	// Will fail because no LLM manager, but validates parsing
	_, err := tool.Execute(context.Background(), params)

	require.Error(t, err)
	// The error should be about LLM manager, not parameter parsing
	assert.Contains(t, err.Error(), "LLM manager not configured")
}

func TestSpawnSubAgentToolIntegration(t *testing.T) {
	t.Run("tool integrates with registry", func(t *testing.T) {
		// Verify the tool implements the Tool interface properly
		tool := NewSpawnSubAgentTool(nil)

		// These should not panic
		name := tool.Name()
		desc := tool.Description()
		params := tool.Parameters()
		reqConf := tool.RequiresConfirmation()

		assert.NotEmpty(t, name)
		assert.NotEmpty(t, desc)
		assert.Equal(t, "object", params.Type)
		assert.False(t, reqConf)
	})
}

func TestSpawnSubAgentTool_ParameterTypes(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)
	params := tool.Parameters()

	t.Run("task_prompt is string type", func(t *testing.T) {
		assert.Equal(t, "string", params.Properties["task_prompt"].Type)
	})

	t.Run("role is string type", func(t *testing.T) {
		assert.Equal(t, "string", params.Properties["role"].Type)
	})

	t.Run("context is string type", func(t *testing.T) {
		assert.Equal(t, "string", params.Properties["context"].Type)
	})

	t.Run("provider is string type", func(t *testing.T) {
		assert.Equal(t, "string", params.Properties["provider"].Type)
	})

	t.Run("model is string type", func(t *testing.T) {
		assert.Equal(t, "string", params.Properties["model"].Type)
	})
}

func TestSpawnSubAgentTool_RoleEnumComplete(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)
	params := tool.Parameters()

	roleProperty := params.Properties["role"]
	expectedRoles := []string{
		"general", "planner", "coder", "reviewer",
		"researcher", "writer", "analyst", "debugger",
		"tester", "synthesizer",
	}

	assert.Len(t, roleProperty.Enum, len(expectedRoles))

	for _, role := range expectedRoles {
		assert.Contains(t, roleProperty.Enum, role, "missing role: %s", role)
	}
}

// Mock types for testing
type mockLLMManager struct {
	llm.Manager
}

func TestSpawnSubAgentTool_ToLLMToolDefinition(t *testing.T) {
	tool := NewSpawnSubAgentTool(nil)

	// Convert to LLM tool definition format
	def := llm.ToolDefinition{
		Name:        tool.Name(),
		Description: tool.Description(),
		Parameters:  tool.Parameters(),
	}

	assert.Equal(t, "spawn_sub_agent", def.Name)
	assert.NotEmpty(t, def.Description)
	assert.Equal(t, "object", def.Parameters.Type)
	assert.Len(t, def.Parameters.Required, 2)
}
