package agent

import (
	"strings"
	"testing"
)

func TestAgentSystemPromptNotEmpty(t *testing.T) {
	if AGENT_SYSTEM_PROMPT == "" {
		t.Error("AGENT_SYSTEM_PROMPT should not be empty")
	}
}

func TestOrchestratorAgentSystemPromptNotEmpty(t *testing.T) {
	if ORCHESTRATOR_AGENT_SYSTEM_PROMPT == "" {
		t.Error("ORCHESTRATOR_AGENT_SYSTEM_PROMPT should not be empty")
	}
}

func TestAgentSystemPromptContainsKeyElements(t *testing.T) {
	keyElements := []string{
		"tool",           // Tool usage guidance
		"security",       // Security considerations
		"file",           // File modification guidance
		"error",          // Error handling
		"code",           // Code-related instructions
	}

	prompt := strings.ToLower(AGENT_SYSTEM_PROMPT)
	for _, element := range keyElements {
		if !strings.Contains(prompt, element) {
			t.Errorf("AGENT_SYSTEM_PROMPT should contain '%s'", element)
		}
	}
}

func TestOrchestratorAgentSystemPromptContainsKeyElements(t *testing.T) {
	keyElements := []string{
		"agent",          // Agent management
		"task",           // Task decomposition
		"coordinate",     // Coordination workflows
		"synthesize",     // Result synthesis
		"parallel",       // Parallel execution
	}

	prompt := strings.ToLower(ORCHESTRATOR_AGENT_SYSTEM_PROMPT)
	for _, element := range keyElements {
		if !strings.Contains(prompt, element) {
			t.Errorf("ORCHESTRATOR_AGENT_SYSTEM_PROMPT should contain '%s'", element)
		}
	}
}

func TestAgentSystemPromptHasMinimumLength(t *testing.T) {
	// A comprehensive prompt should have substantial content
	minLength := 500
	if len(AGENT_SYSTEM_PROMPT) < minLength {
		t.Errorf("AGENT_SYSTEM_PROMPT should be at least %d characters, got %d", minLength, len(AGENT_SYSTEM_PROMPT))
	}
}

func TestOrchestratorAgentSystemPromptHasMinimumLength(t *testing.T) {
	// A comprehensive prompt should have substantial content
	minLength := 500
	if len(ORCHESTRATOR_AGENT_SYSTEM_PROMPT) < minLength {
		t.Errorf("ORCHESTRATOR_AGENT_SYSTEM_PROMPT should be at least %d characters, got %d", minLength, len(ORCHESTRATOR_AGENT_SYSTEM_PROMPT))
	}
}

func TestPromptConcatenationWithContext(t *testing.T) {
	context := "Working on a Go backend project with PostgreSQL database."

	// Simulate how prompts would be concatenated with context
	combinedAgent := AGENT_SYSTEM_PROMPT + "\n\nContext: " + context
	if !strings.Contains(combinedAgent, AGENT_SYSTEM_PROMPT) {
		t.Error("Combined agent prompt should contain original AGENT_SYSTEM_PROMPT")
	}
	if !strings.Contains(combinedAgent, context) {
		t.Error("Combined agent prompt should contain context")
	}

	combinedOrchestrator := ORCHESTRATOR_AGENT_SYSTEM_PROMPT + "\n\nContext: " + context
	if !strings.Contains(combinedOrchestrator, ORCHESTRATOR_AGENT_SYSTEM_PROMPT) {
		t.Error("Combined orchestrator prompt should contain original ORCHESTRATOR_AGENT_SYSTEM_PROMPT")
	}
	if !strings.Contains(combinedOrchestrator, context) {
		t.Error("Combined orchestrator prompt should contain context")
	}
}

func TestPromptsAreDistinct(t *testing.T) {
	if AGENT_SYSTEM_PROMPT == ORCHESTRATOR_AGENT_SYSTEM_PROMPT {
		t.Error("AGENT_SYSTEM_PROMPT and ORCHESTRATOR_AGENT_SYSTEM_PROMPT should be different")
	}
}

func TestAgentPromptContainsSecurityGuidelines(t *testing.T) {
	securityTerms := []string{
		"sensitive",
		"api key",
		"password",
		"untrusted",
	}

	prompt := strings.ToLower(AGENT_SYSTEM_PROMPT)
	foundCount := 0
	for _, term := range securityTerms {
		if strings.Contains(prompt, term) {
			foundCount++
		}
	}

	// Should contain at least some security-related terms
	if foundCount < 2 {
		t.Errorf("AGENT_SYSTEM_PROMPT should contain more security-related guidance, found %d of %d terms", foundCount, len(securityTerms))
	}
}

func TestOrchestratorPromptContainsAgentRoles(t *testing.T) {
	agentRoles := []string{
		"planner",
		"coder",
		"reviewer",
		"researcher",
		"tester",
	}

	prompt := strings.ToLower(ORCHESTRATOR_AGENT_SYSTEM_PROMPT)
	foundCount := 0
	for _, role := range agentRoles {
		if strings.Contains(prompt, role) {
			foundCount++
		}
	}

	// Should mention most of the agent roles
	if foundCount < 4 {
		t.Errorf("ORCHESTRATOR_AGENT_SYSTEM_PROMPT should reference agent roles, found %d of %d roles", foundCount, len(agentRoles))
	}
}

func TestOrchestratorPromptContainsCoordinationStrategies(t *testing.T) {
	strategies := []string{
		"parallel",
		"pipeline",
		"debate",
		"consensus",
	}

	prompt := strings.ToLower(ORCHESTRATOR_AGENT_SYSTEM_PROMPT)
	foundCount := 0
	for _, strategy := range strategies {
		if strings.Contains(prompt, strategy) {
			foundCount++
		}
	}

	// Should mention most coordination strategies
	if foundCount < 3 {
		t.Errorf("ORCHESTRATOR_AGENT_SYSTEM_PROMPT should reference coordination strategies, found %d of %d strategies", foundCount, len(strategies))
	}
}
