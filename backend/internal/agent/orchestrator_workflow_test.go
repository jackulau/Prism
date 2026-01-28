package agent

import (
	"context"
	"testing"
	"time"

	"github.com/jacklau/prism/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrchestratorWorkflow(t *testing.T) {
	t.Run("creates workflow with default config", func(t *testing.T) {
		config := OrchestratorConfig{
			Name:     "test-orchestrator",
			Provider: "anthropic",
			Model:    "claude-sonnet-4-20250514",
		}

		workflow := NewOrchestratorWorkflow(config, nil)

		assert.NotEmpty(t, workflow.ID)
		assert.Equal(t, OrchestratorWorkflowStatusPending, workflow.Status)
		assert.Equal(t, ReasoningEffortHigh, workflow.Config.ReasoningEffort)
		assert.Equal(t, 10, workflow.Config.MaxSubAgents)
		assert.Equal(t, 30*time.Minute, workflow.Config.Timeout)
		assert.Equal(t, 16000, workflow.Config.MaxThinkingTokens)
	})

	t.Run("creates workflow with custom config", func(t *testing.T) {
		config := OrchestratorConfig{
			ID:              "custom-id",
			Name:            "custom-orchestrator",
			Provider:        "openai",
			Model:           "gpt-4",
			ReasoningEffort: ReasoningEffortMedium,
			MaxSubAgents:    5,
			Timeout:         10 * time.Minute,
			VerboseLogging:  true,
		}

		workflow := NewOrchestratorWorkflow(config, nil)

		assert.Equal(t, "custom-id", workflow.ID)
		assert.Equal(t, ReasoningEffortMedium, workflow.Config.ReasoningEffort)
		assert.Equal(t, 5, workflow.Config.MaxSubAgents)
		assert.Equal(t, 10*time.Minute, workflow.Config.Timeout)
		assert.True(t, workflow.Config.VerboseLogging)
	})

	t.Run("creates workflow with custom system prompt", func(t *testing.T) {
		customPrompt := "You are a custom orchestrator."
		config := OrchestratorConfig{
			Name:         "custom-prompt-orchestrator",
			Provider:     "anthropic",
			Model:        "claude-sonnet-4-20250514",
			SystemPrompt: customPrompt,
		}

		workflow := NewOrchestratorWorkflow(config, nil)

		assert.Equal(t, customPrompt, workflow.agent.Config.SystemPrompt)
	})
}

func TestOrchestratorWorkflow_ExtractImages(t *testing.T) {
	t.Run("extracts images from messages", func(t *testing.T) {
		config := OrchestratorConfig{
			Name:           "test-orchestrator",
			Provider:       "anthropic",
			Model:          "claude-sonnet-4-20250514",
			VerboseLogging: true,
		}

		workflow := NewOrchestratorWorkflow(config, nil)

		messages := []llm.Message{
			{
				Role:    "user",
				Content: "Here is an image",
				Images: []llm.ImageData{
					{
						URL:      "https://example.com/image1.png",
						MimeType: "image/png",
					},
				},
			},
			{
				Role:    "assistant",
				Content: "I see the image",
			},
			{
				Role:    "user",
				Content: "Another image",
				Images: []llm.ImageData{
					{
						Base64:   "base64encodeddata",
						MimeType: "image/jpeg",
					},
				},
			},
		}

		images := workflow.ExtractImages(messages)

		require.Len(t, images, 2)

		assert.Equal(t, "https://example.com/image1.png", images[0].URL)
		assert.Equal(t, "image/png", images[0].MimeType)
		assert.Equal(t, "user", images[0].Source)
		assert.Equal(t, 0, images[0].MessageIndex)

		assert.Equal(t, "base64encodeddata", images[1].Base64)
		assert.Equal(t, "image/jpeg", images[1].MimeType)
		assert.Equal(t, "user", images[1].Source)
		assert.Equal(t, 2, images[1].MessageIndex)
	})

	t.Run("returns empty slice for messages without images", func(t *testing.T) {
		config := OrchestratorConfig{
			Name:     "test-orchestrator",
			Provider: "anthropic",
			Model:    "claude-sonnet-4-20250514",
		}

		workflow := NewOrchestratorWorkflow(config, nil)

		messages := []llm.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there"},
		}

		images := workflow.ExtractImages(messages)

		assert.Len(t, images, 0)
	})
}

func TestOrchestratorWorkflow_ReasoningLogs(t *testing.T) {
	t.Run("adds and retrieves reasoning logs", func(t *testing.T) {
		config := OrchestratorConfig{
			Name:           "test-orchestrator",
			Provider:       "anthropic",
			Model:          "claude-sonnet-4-20250514",
			VerboseLogging: true,
		}

		workflow := NewOrchestratorWorkflow(config, nil)

		// Add reasoning logs via the internal method
		workflow.addReasoningLog("Decision 1", "Reasoning for decision 1", "")
		workflow.addReasoningLog("Decision 2", "Reasoning for decision 2", "agent-1")

		logs := workflow.GetReasoningLogs()

		require.Len(t, logs, 2)
		assert.Equal(t, "Decision 1", logs[0].Decision)
		assert.Equal(t, "Reasoning for decision 1", logs[0].Reasoning)
		assert.Empty(t, logs[0].SubAgentID)

		assert.Equal(t, "Decision 2", logs[1].Decision)
		assert.Equal(t, "Reasoning for decision 2", logs[1].Reasoning)
		assert.Equal(t, "agent-1", logs[1].SubAgentID)
	})
}

func TestOrchestratorWorkflow_Events(t *testing.T) {
	t.Run("emits events when verbose logging is enabled", func(t *testing.T) {
		config := OrchestratorConfig{
			Name:           "test-orchestrator",
			Provider:       "anthropic",
			Model:          "claude-sonnet-4-20250514",
			VerboseLogging: true,
		}

		workflow := NewOrchestratorWorkflow(config, nil)

		// Emit an event
		workflow.emitEvent(OrchestratorEventStarted, "", map[string]interface{}{
			"test": "data",
		})

		// Check event was received
		select {
		case event := <-workflow.Events():
			assert.Equal(t, OrchestratorEventStarted, event.Type)
			assert.Equal(t, workflow.ID, event.WorkflowID)
			assert.Equal(t, "data", event.Data["test"])
		case <-time.After(100 * time.Millisecond):
			t.Fatal("expected event but none received")
		}
	})
}

func TestOrchestratorWorkflow_Stop(t *testing.T) {
	t.Run("cancels workflow and sub-agents", func(t *testing.T) {
		config := OrchestratorConfig{
			Name:     "test-orchestrator",
			Provider: "anthropic",
			Model:    "claude-sonnet-4-20250514",
		}

		workflow := NewOrchestratorWorkflow(config, nil)

		// Add a mock sub-agent
		subAgentConfig := AgentConfig{
			ID:       "sub-agent-1",
			Provider: "anthropic",
			Model:    "claude-sonnet-4-20250514",
		}
		subAgent := NewAgent(subAgentConfig, nil)
		workflow.subAgents = append(workflow.subAgents, subAgent)

		// Stop the workflow
		workflow.Stop()

		assert.Equal(t, OrchestratorWorkflowStatusCancelled, workflow.Status)
	})
}

func TestOrchestratorWorkflow_AddMessage(t *testing.T) {
	t.Run("adds messages to conversation history", func(t *testing.T) {
		config := OrchestratorConfig{
			Name:     "test-orchestrator",
			Provider: "anthropic",
			Model:    "claude-sonnet-4-20250514",
		}

		workflow := NewOrchestratorWorkflow(config, nil)

		msg1 := llm.Message{Role: "user", Content: "Hello"}
		msg2 := llm.Message{Role: "assistant", Content: "Hi there"}

		workflow.AddMessage(msg1)
		workflow.AddMessage(msg2)

		assert.Len(t, workflow.messages, 2)
		assert.Equal(t, "Hello", workflow.messages[0].Content)
		assert.Equal(t, "Hi there", workflow.messages[1].Content)
	})
}

func TestOrchestratorWorkflow_GetSubAgents(t *testing.T) {
	t.Run("returns copy of sub-agents", func(t *testing.T) {
		config := OrchestratorConfig{
			Name:     "test-orchestrator",
			Provider: "anthropic",
			Model:    "claude-sonnet-4-20250514",
		}

		workflow := NewOrchestratorWorkflow(config, nil)

		// Add sub-agents
		for i := 0; i < 3; i++ {
			subAgentConfig := AgentConfig{
				ID:       "sub-agent-" + string(rune('1'+i)),
				Provider: "anthropic",
				Model:    "claude-sonnet-4-20250514",
			}
			workflow.subAgents = append(workflow.subAgents, NewAgent(subAgentConfig, nil))
		}

		subAgents := workflow.GetSubAgents()

		assert.Len(t, subAgents, 3)
	})
}

func TestOrchestratorWorkflow_Run(t *testing.T) {
	t.Run("fails if already running", func(t *testing.T) {
		config := OrchestratorConfig{
			Name:     "test-orchestrator",
			Provider: "anthropic",
			Model:    "claude-sonnet-4-20250514",
		}

		workflow := NewOrchestratorWorkflow(config, nil)
		workflow.Status = OrchestratorWorkflowStatusRunning

		task := NewTask("Test task")
		err := workflow.Run(context.Background(), task)

		assert.Equal(t, ErrAgentAlreadyRunning, err)
	})
}

func TestOrchestratorWorkflow_SpawnSubAgent(t *testing.T) {
	t.Run("fails when max sub-agents reached", func(t *testing.T) {
		config := OrchestratorConfig{
			Name:         "test-orchestrator",
			Provider:     "anthropic",
			Model:        "claude-sonnet-4-20250514",
			MaxSubAgents: 1,
		}

		workflow := NewOrchestratorWorkflow(config, nil)

		// Add a sub-agent to reach max
		subAgentConfig := AgentConfig{
			ID:       "sub-agent-1",
			Provider: "anthropic",
			Model:    "claude-sonnet-4-20250514",
		}
		workflow.subAgents = append(workflow.subAgents, NewAgent(subAgentConfig, nil))

		// Try to spawn another
		task := NewTask("Test task")
		_, err := workflow.SpawnSubAgent(AgentConfig{ID: "sub-agent-2"}, task)

		assert.Equal(t, ErrTooManyAgents, err)
	})
}

func TestReasoningEffort(t *testing.T) {
	t.Run("reasoning effort constants are correct", func(t *testing.T) {
		assert.Equal(t, ReasoningEffort("low"), ReasoningEffortLow)
		assert.Equal(t, ReasoningEffort("medium"), ReasoningEffortMedium)
		assert.Equal(t, ReasoningEffort("high"), ReasoningEffortHigh)
	})
}

func TestOrchestratorWorkflowStatus(t *testing.T) {
	t.Run("status constants are correct", func(t *testing.T) {
		assert.Equal(t, OrchestratorWorkflowStatus("pending"), OrchestratorWorkflowStatusPending)
		assert.Equal(t, OrchestratorWorkflowStatus("running"), OrchestratorWorkflowStatusRunning)
		assert.Equal(t, OrchestratorWorkflowStatus("completed"), OrchestratorWorkflowStatusCompleted)
		assert.Equal(t, OrchestratorWorkflowStatus("failed"), OrchestratorWorkflowStatusFailed)
		assert.Equal(t, OrchestratorWorkflowStatus("cancelled"), OrchestratorWorkflowStatusCancelled)
	})
}

func TestOrchestratorEventType(t *testing.T) {
	t.Run("event type constants are correct", func(t *testing.T) {
		assert.Equal(t, OrchestratorEventType("started"), OrchestratorEventStarted)
		assert.Equal(t, OrchestratorEventType("reasoning"), OrchestratorEventReasoning)
		assert.Equal(t, OrchestratorEventType("spawning_agent"), OrchestratorEventSpawningAgent)
		assert.Equal(t, OrchestratorEventType("agent_completed"), OrchestratorEventAgentCompleted)
		assert.Equal(t, OrchestratorEventType("agent_failed"), OrchestratorEventAgentFailed)
		assert.Equal(t, OrchestratorEventType("completed"), OrchestratorEventCompleted)
		assert.Equal(t, OrchestratorEventType("failed"), OrchestratorEventFailed)
		assert.Equal(t, OrchestratorEventType("cancelled"), OrchestratorEventCancelled)
	})
}

func TestBuildOrchestratorSystemPrompt(t *testing.T) {
	t.Run("returns non-empty system prompt", func(t *testing.T) {
		prompt := buildOrchestratorSystemPrompt()

		assert.NotEmpty(t, prompt)
		assert.Contains(t, prompt, "orchestrator")
		assert.Contains(t, prompt, "sub-agent")
		assert.Contains(t, prompt, "spawn_sub_agent")
	})
}

func TestImageContent(t *testing.T) {
	t.Run("image content struct works correctly", func(t *testing.T) {
		img := ImageContent{
			ID:           "test-id",
			URL:          "https://example.com/image.png",
			Base64:       "",
			MimeType:     "image/png",
			Source:       "user",
			MessageIndex: 0,
		}

		assert.Equal(t, "test-id", img.ID)
		assert.Equal(t, "https://example.com/image.png", img.URL)
		assert.Equal(t, "image/png", img.MimeType)
		assert.Equal(t, "user", img.Source)
		assert.Equal(t, 0, img.MessageIndex)
	})
}

func TestReasoningLog(t *testing.T) {
	t.Run("reasoning log struct works correctly", func(t *testing.T) {
		now := time.Now()
		log := ReasoningLog{
			ID:         "test-id",
			Timestamp:  now,
			Decision:   "Test decision",
			Reasoning:  "Test reasoning",
			SubAgentID: "agent-1",
			Metadata:   map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "test-id", log.ID)
		assert.Equal(t, now, log.Timestamp)
		assert.Equal(t, "Test decision", log.Decision)
		assert.Equal(t, "Test reasoning", log.Reasoning)
		assert.Equal(t, "agent-1", log.SubAgentID)
		assert.Equal(t, "value", log.Metadata["key"])
	})
}

func TestSubAgentResult(t *testing.T) {
	t.Run("sub-agent result struct works correctly", func(t *testing.T) {
		result := SubAgentResult{
			AgentID:  "agent-1",
			Role:     RoleCoder,
			Task:     "Write code",
			Output:   "Code output",
			Success:  true,
			Duration: 5 * time.Second,
		}

		assert.Equal(t, "agent-1", result.AgentID)
		assert.Equal(t, RoleCoder, result.Role)
		assert.Equal(t, "Write code", result.Task)
		assert.Equal(t, "Code output", result.Output)
		assert.True(t, result.Success)
		assert.Equal(t, 5*time.Second, result.Duration)
	})
}

func TestOrchestratorResult(t *testing.T) {
	t.Run("orchestrator result struct works correctly", func(t *testing.T) {
		now := time.Now()
		result := OrchestratorResult{
			ID:              "result-1",
			Success:         true,
			Output:          "Final output",
			SubAgentResults: []SubAgentResult{},
			ReasoningLogs:   []ReasoningLog{},
			Images:          []ImageContent{},
			Duration:        10 * time.Second,
			CompletedAt:     now,
		}

		assert.Equal(t, "result-1", result.ID)
		assert.True(t, result.Success)
		assert.Equal(t, "Final output", result.Output)
		assert.Equal(t, 10*time.Second, result.Duration)
		assert.Equal(t, now, result.CompletedAt)
	})
}
