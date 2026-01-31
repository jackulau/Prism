package websocket

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewAgentProgress(t *testing.T) {
	progress := &AgentProgressInfo{
		CurrentStep:     2,
		TotalSteps:      5,
		PercentComplete: 40.0,
		StepName:        "test_step",
		Message:         "Processing...",
		Timestamp:       time.Now().UnixMilli(),
	}

	msg := NewAgentProgress("agent-123", progress)

	if msg.Type != TypeAgentProgress {
		t.Errorf("expected type %s, got %s", TypeAgentProgress, msg.Type)
	}
	if msg.AgentID != "agent-123" {
		t.Errorf("expected agent_id agent-123, got %s", msg.AgentID)
	}
	if msg.CurrentStep != 2 {
		t.Errorf("expected current_step 2, got %d", msg.CurrentStep)
	}
	if msg.TotalSteps != 5 {
		t.Errorf("expected total_steps 5, got %d", msg.TotalSteps)
	}
	if msg.StepName != "test_step" {
		t.Errorf("expected step_name test_step, got %s", msg.StepName)
	}
	if msg.Message != "Processing..." {
		t.Errorf("expected message 'Processing...', got %s", msg.Message)
	}

	// Verify metadata contains percent_complete
	if msg.Metadata == nil {
		t.Error("expected metadata to be set")
	}
	if pct, ok := msg.Metadata["percent_complete"].(float64); !ok || pct != 40.0 {
		t.Errorf("expected percent_complete 40.0, got %v", msg.Metadata["percent_complete"])
	}
}

func TestNewAgentStepStarted(t *testing.T) {
	step := &AgentStepInfo{
		StepID:    "step-1",
		StepName:  "run_llm",
		StepType:  "llm",
		Status:    "running",
		StartedAt: time.Now().UnixMilli(),
		Input:     "test input",
	}

	msg := NewAgentStepStarted("agent-123", step)

	if msg.Type != TypeAgentStepStarted {
		t.Errorf("expected type %s, got %s", TypeAgentStepStarted, msg.Type)
	}
	if msg.AgentID != "agent-123" {
		t.Errorf("expected agent_id agent-123, got %s", msg.AgentID)
	}
	if msg.StepID != "step-1" {
		t.Errorf("expected step_id step-1, got %s", msg.StepID)
	}
	if msg.StepName != "run_llm" {
		t.Errorf("expected step_name run_llm, got %s", msg.StepName)
	}
	if msg.StepType != "llm" {
		t.Errorf("expected step_type llm, got %s", msg.StepType)
	}
	if msg.Status != "running" {
		t.Errorf("expected status running, got %s", msg.Status)
	}
}

func TestNewAgentStepCompleted(t *testing.T) {
	step := &AgentStepInfo{
		StepID:      "step-1",
		StepName:    "run_llm",
		StepType:    "llm",
		Status:      "completed",
		Output:      "test output",
		Duration:    1500,
		StartedAt:   time.Now().Add(-2 * time.Second).UnixMilli(),
		CompletedAt: time.Now().UnixMilli(),
	}

	msg := NewAgentStepCompleted("agent-123", step)

	if msg.Type != TypeAgentStepCompleted {
		t.Errorf("expected type %s, got %s", TypeAgentStepCompleted, msg.Type)
	}
	if msg.AgentID != "agent-123" {
		t.Errorf("expected agent_id agent-123, got %s", msg.AgentID)
	}
	if msg.Duration != 1500 {
		t.Errorf("expected duration 1500, got %d", msg.Duration)
	}
	if msg.Status != "completed" {
		t.Errorf("expected status completed, got %s", msg.Status)
	}
}

func TestNewAgentThinkingStart(t *testing.T) {
	timestamp := time.Now().UnixMilli()
	msg := NewAgentThinkingStart("agent-123", "Analyzing the code structure", timestamp)

	if msg.Type != TypeAgentThinkingStart {
		t.Errorf("expected type %s, got %s", TypeAgentThinkingStart, msg.Type)
	}
	if msg.AgentID != "agent-123" {
		t.Errorf("expected agent_id agent-123, got %s", msg.AgentID)
	}
	if msg.Message != "Analyzing the code structure" {
		t.Errorf("expected message 'Analyzing the code structure', got %s", msg.Message)
	}

	// Verify metadata
	if msg.Metadata == nil {
		t.Error("expected metadata to be set")
	}
	if isThinking, ok := msg.Metadata["is_thinking"].(bool); !ok || !isThinking {
		t.Errorf("expected is_thinking true, got %v", msg.Metadata["is_thinking"])
	}
}

func TestNewAgentThinkingEnd(t *testing.T) {
	timestamp := time.Now().UnixMilli()
	msg := NewAgentThinkingEnd("agent-123", timestamp)

	if msg.Type != TypeAgentThinkingEnd {
		t.Errorf("expected type %s, got %s", TypeAgentThinkingEnd, msg.Type)
	}
	if msg.AgentID != "agent-123" {
		t.Errorf("expected agent_id agent-123, got %s", msg.AgentID)
	}

	// Verify metadata
	if msg.Metadata == nil {
		t.Error("expected metadata to be set")
	}
	if isThinking, ok := msg.Metadata["is_thinking"].(bool); !ok || isThinking {
		t.Errorf("expected is_thinking false, got %v", msg.Metadata["is_thinking"])
	}
}

func TestNewAgentEstimate(t *testing.T) {
	estimate := &AgentEstimateInfo{
		EstimatedSteps:     10,
		EstimatedDurationMs: 30000,
		Confidence:         0.8,
		Message:            "Estimated 10 steps",
		Timestamp:          time.Now().UnixMilli(),
	}

	msg := NewAgentEstimate("agent-123", estimate)

	if msg.Type != TypeAgentEstimate {
		t.Errorf("expected type %s, got %s", TypeAgentEstimate, msg.Type)
	}
	if msg.AgentID != "agent-123" {
		t.Errorf("expected agent_id agent-123, got %s", msg.AgentID)
	}
	if msg.TotalSteps != 10 {
		t.Errorf("expected total_steps 10, got %d", msg.TotalSteps)
	}
	if msg.Message != "Estimated 10 steps" {
		t.Errorf("expected message 'Estimated 10 steps', got %s", msg.Message)
	}

	// Verify metadata
	if msg.Metadata == nil {
		t.Error("expected metadata to be set")
	}
	if confidence, ok := msg.Metadata["confidence"].(float64); !ok || confidence != 0.8 {
		t.Errorf("expected confidence 0.8, got %v", msg.Metadata["confidence"])
	}
}

func TestProgressMessageJSONSerialization(t *testing.T) {
	progress := &AgentProgressInfo{
		CurrentStep:     3,
		TotalSteps:      10,
		PercentComplete: 30.0,
		StepName:        "processing",
		Message:         "Working on task",
		Timestamp:       1704067200000,
	}

	msg := NewAgentProgress("agent-456", progress)

	// Serialize to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	// Deserialize back
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	// Verify key fields
	if result["type"] != TypeAgentProgress {
		t.Errorf("expected type %s in JSON, got %v", TypeAgentProgress, result["type"])
	}
	if result["agent_id"] != "agent-456" {
		t.Errorf("expected agent_id agent-456 in JSON, got %v", result["agent_id"])
	}
	if result["current_step"].(float64) != 3 {
		t.Errorf("expected current_step 3 in JSON, got %v", result["current_step"])
	}
	if result["total_steps"].(float64) != 10 {
		t.Errorf("expected total_steps 10 in JSON, got %v", result["total_steps"])
	}
}

func TestStepInfoJSONSerialization(t *testing.T) {
	step := &AgentStepInfo{
		StepID:      "step-abc",
		StepName:    "create_sandbox",
		StepType:    "sandbox",
		Status:      "completed",
		Output:      "Sandbox created successfully",
		Duration:    2500,
		StartedAt:   1704067200000,
		CompletedAt: 1704067202500,
		Metadata: map[string]interface{}{
			"sandbox_id": "sb-123",
		},
	}

	msg := NewAgentStepCompleted("agent-789", step)

	// Serialize to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	// Deserialize back
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	// Verify key fields
	if result["type"] != TypeAgentStepCompleted {
		t.Errorf("expected type %s in JSON, got %v", TypeAgentStepCompleted, result["type"])
	}
	if result["step_id"] != "step-abc" {
		t.Errorf("expected step_id step-abc in JSON, got %v", result["step_id"])
	}
	if result["step_name"] != "create_sandbox" {
		t.Errorf("expected step_name create_sandbox in JSON, got %v", result["step_name"])
	}
	if result["duration"].(float64) != 2500 {
		t.Errorf("expected duration 2500 in JSON, got %v", result["duration"])
	}
}

func TestAgentProgressInfoStruct(t *testing.T) {
	info := AgentProgressInfo{
		CurrentStep:     5,
		TotalSteps:      10,
		PercentComplete: 50.0,
		StepName:        "run_llm",
		Message:         "Processing request",
		Timestamp:       1704067200000,
	}

	// Test JSON marshaling
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal AgentProgressInfo: %v", err)
	}

	var result AgentProgressInfo
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal AgentProgressInfo: %v", err)
	}

	if result.CurrentStep != 5 {
		t.Errorf("expected current_step 5, got %d", result.CurrentStep)
	}
	if result.TotalSteps != 10 {
		t.Errorf("expected total_steps 10, got %d", result.TotalSteps)
	}
	if result.PercentComplete != 50.0 {
		t.Errorf("expected percent_complete 50.0, got %f", result.PercentComplete)
	}
}

func TestAgentThinkingInfoStruct(t *testing.T) {
	info := AgentThinkingInfo{
		IsThinking: true,
		Context:    "Analyzing code",
		Timestamp:  1704067200000,
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal AgentThinkingInfo: %v", err)
	}

	var result AgentThinkingInfo
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal AgentThinkingInfo: %v", err)
	}

	if !result.IsThinking {
		t.Error("expected is_thinking true")
	}
	if result.Context != "Analyzing code" {
		t.Errorf("expected context 'Analyzing code', got %s", result.Context)
	}
}

func TestAgentEstimateInfoStruct(t *testing.T) {
	info := AgentEstimateInfo{
		EstimatedSteps:     15,
		EstimatedDurationMs: 45000,
		Confidence:         0.75,
		Message:            "Estimated completion time",
		Timestamp:          1704067200000,
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal AgentEstimateInfo: %v", err)
	}

	var result AgentEstimateInfo
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal AgentEstimateInfo: %v", err)
	}

	if result.EstimatedSteps != 15 {
		t.Errorf("expected estimated_steps 15, got %d", result.EstimatedSteps)
	}
	if result.EstimatedDurationMs != 45000 {
		t.Errorf("expected estimated_duration_ms 45000, got %d", result.EstimatedDurationMs)
	}
	if result.Confidence != 0.75 {
		t.Errorf("expected confidence 0.75, got %f", result.Confidence)
	}
}
