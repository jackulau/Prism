package agent

import (
	"encoding/json"
	"testing"
	"time"
)

func TestProgressData_JSONSerialization(t *testing.T) {
	data := ProgressData{
		CurrentStep:     2,
		TotalSteps:      5,
		PercentComplete: 40.0,
		StepName:        "processing",
		Message:         "Processing data",
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal ProgressData: %v", err)
	}

	var decoded ProgressData
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ProgressData: %v", err)
	}

	if decoded.CurrentStep != data.CurrentStep {
		t.Errorf("CurrentStep mismatch: got %d, want %d", decoded.CurrentStep, data.CurrentStep)
	}
	if decoded.TotalSteps != data.TotalSteps {
		t.Errorf("TotalSteps mismatch: got %d, want %d", decoded.TotalSteps, data.TotalSteps)
	}
	if decoded.PercentComplete != data.PercentComplete {
		t.Errorf("PercentComplete mismatch: got %f, want %f", decoded.PercentComplete, data.PercentComplete)
	}
	if decoded.StepName != data.StepName {
		t.Errorf("StepName mismatch: got %s, want %s", decoded.StepName, data.StepName)
	}
	if decoded.Message != data.Message {
		t.Errorf("Message mismatch: got %s, want %s", decoded.Message, data.Message)
	}
}

func TestEstimateData_JSONSerialization(t *testing.T) {
	data := EstimateData{
		EstimatedTokensRemaining: 500,
		EstimatedTimeMs:          30000,
		Confidence:               0.85,
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal EstimateData: %v", err)
	}

	var decoded EstimateData
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal EstimateData: %v", err)
	}

	if decoded.EstimatedTokensRemaining != data.EstimatedTokensRemaining {
		t.Errorf("EstimatedTokensRemaining mismatch: got %d, want %d", decoded.EstimatedTokensRemaining, data.EstimatedTokensRemaining)
	}
	if decoded.EstimatedTimeMs != data.EstimatedTimeMs {
		t.Errorf("EstimatedTimeMs mismatch: got %d, want %d", decoded.EstimatedTimeMs, data.EstimatedTimeMs)
	}
	if decoded.Confidence != data.Confidence {
		t.Errorf("Confidence mismatch: got %f, want %f", decoded.Confidence, data.Confidence)
	}
}

func TestStepData_JSONSerialization(t *testing.T) {
	data := StepData{
		StepNumber:  3,
		TotalSteps:  10,
		StepName:    "validation",
		Description: "Validating input parameters",
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal StepData: %v", err)
	}

	var decoded StepData
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal StepData: %v", err)
	}

	if decoded.StepNumber != data.StepNumber {
		t.Errorf("StepNumber mismatch: got %d, want %d", decoded.StepNumber, data.StepNumber)
	}
	if decoded.TotalSteps != data.TotalSteps {
		t.Errorf("TotalSteps mismatch: got %d, want %d", decoded.TotalSteps, data.TotalSteps)
	}
	if decoded.StepName != data.StepName {
		t.Errorf("StepName mismatch: got %s, want %s", decoded.StepName, data.StepName)
	}
	if decoded.Description != data.Description {
		t.Errorf("Description mismatch: got %s, want %s", decoded.Description, data.Description)
	}
}

func TestThinkingData_JSONSerialization(t *testing.T) {
	data := ThinkingData{
		Phase:       "reasoning",
		Description: "Analyzing the problem",
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal ThinkingData: %v", err)
	}

	var decoded ThinkingData
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ThinkingData: %v", err)
	}

	if decoded.Phase != data.Phase {
		t.Errorf("Phase mismatch: got %s, want %s", decoded.Phase, data.Phase)
	}
	if decoded.Description != data.Description {
		t.Errorf("Description mismatch: got %s, want %s", decoded.Description, data.Description)
	}
}

func TestAgentEventType_Constants(t *testing.T) {
	tests := []struct {
		eventType AgentEventType
		expected  string
	}{
		{AgentEventProgress, "progress"},
		{AgentEventStepStarted, "step_started"},
		{AgentEventStepCompleted, "step_completed"},
		{AgentEventThinkingStart, "thinking_start"},
		{AgentEventThinkingEnd, "thinking_end"},
		{AgentEventEstimate, "estimate"},
	}

	for _, tt := range tests {
		if string(tt.eventType) != tt.expected {
			t.Errorf("AgentEventType constant mismatch: got %s, want %s", tt.eventType, tt.expected)
		}
	}
}

func TestSwarmEventType_Constants(t *testing.T) {
	tests := []struct {
		eventType SwarmEventType
		expected  string
	}{
		{SwarmEventStepProgress, "step_progress"},
		{SwarmEventAgentThinking, "agent_thinking"},
		{SwarmEventEstimate, "estimate"},
	}

	for _, tt := range tests {
		if string(tt.eventType) != tt.expected {
			t.Errorf("SwarmEventType constant mismatch: got %s, want %s", tt.eventType, tt.expected)
		}
	}
}

func TestSwarmProgressData_JSONSerialization(t *testing.T) {
	data := SwarmProgressData{
		CompletedAgents: 3,
		TotalAgents:     5,
		PercentComplete: 60.0,
		CurrentPhase:    "execution",
		Message:         "Processing agents",
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal SwarmProgressData: %v", err)
	}

	var decoded SwarmProgressData
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal SwarmProgressData: %v", err)
	}

	if decoded.CompletedAgents != data.CompletedAgents {
		t.Errorf("CompletedAgents mismatch: got %d, want %d", decoded.CompletedAgents, data.CompletedAgents)
	}
	if decoded.TotalAgents != data.TotalAgents {
		t.Errorf("TotalAgents mismatch: got %d, want %d", decoded.TotalAgents, data.TotalAgents)
	}
	if decoded.PercentComplete != data.PercentComplete {
		t.Errorf("PercentComplete mismatch: got %f, want %f", decoded.PercentComplete, data.PercentComplete)
	}
	if decoded.CurrentPhase != data.CurrentPhase {
		t.Errorf("CurrentPhase mismatch: got %s, want %s", decoded.CurrentPhase, data.CurrentPhase)
	}
}

func TestProgressTracker_EmitsEvents(t *testing.T) {
	events := make(chan *AgentEvent, 100)
	tracker := NewProgressTracker("agent-123", 5, events)

	// Test StartStep
	tracker.StartStep(1, "step1", "First step")

	select {
	case event := <-events:
		if event.Type != AgentEventStepStarted {
			t.Errorf("Expected step_started event, got %s", event.Type)
		}
		if event.AgentID != "agent-123" {
			t.Errorf("Expected agent ID 'agent-123', got %s", event.AgentID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for step_started event")
	}

	// Should also emit a progress event
	select {
	case event := <-events:
		if event.Type != AgentEventProgress {
			t.Errorf("Expected progress event, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for progress event")
	}

	// Test CompleteStep
	tracker.CompleteStep("Step completed successfully")

	select {
	case event := <-events:
		if event.Type != AgentEventStepCompleted {
			t.Errorf("Expected step_completed event, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for step_completed event")
	}

	// Test StartThinking
	tracker.StartThinking("inference", "Processing with LLM")

	select {
	case event := <-events:
		if event.Type != AgentEventThinkingStart {
			t.Errorf("Expected thinking_start event, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for thinking_start event")
	}

	// Test EndThinking
	tracker.EndThinking("inference")

	select {
	case event := <-events:
		if event.Type != AgentEventThinkingEnd {
			t.Errorf("Expected thinking_end event, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for thinking_end event")
	}

	// Test EmitEstimate
	tracker.EmitEstimate(500, 30000, 0.85)

	select {
	case event := <-events:
		if event.Type != AgentEventEstimate {
			t.Errorf("Expected estimate event, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for estimate event")
	}
}

func TestSwarmProgressTracker_EmitsEvents(t *testing.T) {
	events := make(chan *SwarmEvent, 100)
	tracker := NewSwarmProgressTracker("swarm-456", 3, events)

	// Test SetPhase and UpdateProgress
	tracker.SetPhase("execution")
	tracker.UpdateProgress("Starting execution")

	select {
	case event := <-events:
		if event.Type != SwarmEventProgress {
			t.Errorf("Expected progress event, got %s", event.Type)
		}
		if event.SwarmID != "swarm-456" {
			t.Errorf("Expected swarm ID 'swarm-456', got %s", event.SwarmID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for progress event")
	}

	// Test IncrementCompleted
	tracker.IncrementCompleted()

	select {
	case event := <-events:
		if event.Type != SwarmEventProgress {
			t.Errorf("Expected progress event, got %s", event.Type)
		}
		completedAgents, ok := event.Data["completed_agents"].(int)
		if !ok || completedAgents != 1 {
			t.Errorf("Expected completed_agents to be 1, got %v", event.Data["completed_agents"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for progress event")
	}

	// Test StartStep
	tracker.SetTotalSteps(5)
	tracker.StartStep(2, "step2")

	select {
	case event := <-events:
		if event.Type != SwarmEventStepProgress {
			t.Errorf("Expected step_progress event, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for step_progress event")
	}

	// Test EmitAgentThinking
	tracker.EmitAgentThinking("agent-1", RoleCoder, true)

	select {
	case event := <-events:
		if event.Type != SwarmEventAgentThinking {
			t.Errorf("Expected agent_thinking event, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for agent_thinking event")
	}

	// Test EmitEstimate
	tracker.EmitEstimate(60000, 0.75)

	select {
	case event := <-events:
		if event.Type != SwarmEventEstimate {
			t.Errorf("Expected estimate event, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for estimate event")
	}
}

func TestProgressTracker_NilChannel(t *testing.T) {
	// Should not panic with nil events channel
	tracker := NewProgressTracker("agent-123", 5, nil)

	// These should not panic
	tracker.StartStep(1, "step1", "First step")
	tracker.CompleteStep("Done")
	tracker.UpdateProgress("Progress")
	tracker.StartThinking("inference", "Thinking")
	tracker.EndThinking("inference")
	tracker.EmitEstimate(100, 1000, 0.5)
}

func TestSwarmProgressTracker_NilChannel(t *testing.T) {
	// Should not panic with nil events channel
	tracker := NewSwarmProgressTracker("swarm-456", 3, nil)

	// These should not panic
	tracker.SetPhase("execution")
	tracker.SetTotalSteps(5)
	tracker.IncrementCompleted()
	tracker.StartStep(1, "step1")
	tracker.UpdateProgress("Progress")
	tracker.EmitAgentThinking("agent-1", RoleCoder, true)
	tracker.EmitEstimate(1000, 0.5)
}

func TestProgressTracker_PercentCalculation(t *testing.T) {
	events := make(chan *AgentEvent, 100)
	tracker := NewProgressTracker("agent-123", 4, events)

	tracker.StartStep(2, "step2", "Second step")

	// Skip step_started event
	<-events

	// Check progress event
	event := <-events
	if event.Type != AgentEventProgress {
		t.Errorf("Expected progress event, got %s", event.Type)
	}

	percentComplete, ok := event.Data["percent_complete"].(float64)
	if !ok {
		t.Fatal("percent_complete not found in event data")
	}

	// 2/4 = 50%
	expected := 50.0
	if percentComplete != expected {
		t.Errorf("Expected percent_complete to be %f, got %f", expected, percentComplete)
	}
}

func TestSwarmProgressTracker_PercentCalculation(t *testing.T) {
	events := make(chan *SwarmEvent, 100)
	tracker := NewSwarmProgressTracker("swarm-456", 4, events)

	tracker.SetPhase("execution")
	tracker.IncrementCompleted()
	tracker.IncrementCompleted()

	// Drain first event
	<-events

	// Check second event
	event := <-events
	if event.Type != SwarmEventProgress {
		t.Errorf("Expected progress event, got %s", event.Type)
	}

	percentComplete, ok := event.Data["percent_complete"].(float64)
	if !ok {
		t.Fatal("percent_complete not found in event data")
	}

	// 2/4 = 50%
	expected := 50.0
	if percentComplete != expected {
		t.Errorf("Expected percent_complete to be %f, got %f", expected, percentComplete)
	}
}
