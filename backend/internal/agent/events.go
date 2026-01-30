package agent

import "time"

// Extended AgentEventType constants for progress tracking
const (
	// AgentEventProgress indicates a progress update with percentage and step info
	AgentEventProgress AgentEventType = "progress"

	// AgentEventStepStarted indicates a step in a multi-step process has started
	AgentEventStepStarted AgentEventType = "step_started"

	// AgentEventStepCompleted indicates a step has completed successfully
	AgentEventStepCompleted AgentEventType = "step_completed"

	// AgentEventThinkingStart indicates the thinking/inference phase has begun
	AgentEventThinkingStart AgentEventType = "thinking_start"

	// AgentEventThinkingEnd indicates the thinking/inference phase has ended
	AgentEventThinkingEnd AgentEventType = "thinking_end"

	// AgentEventEstimate provides estimated time/tokens remaining
	AgentEventEstimate AgentEventType = "estimate"
)

// ProgressData contains information about execution progress
type ProgressData struct {
	CurrentStep     int     `json:"current_step"`
	TotalSteps      int     `json:"total_steps"`
	PercentComplete float64 `json:"percent_complete"`
	StepName        string  `json:"step_name"`
	Message         string  `json:"message"`
}

// EstimateData contains estimated remaining time/tokens
type EstimateData struct {
	EstimatedTokensRemaining int     `json:"estimated_tokens_remaining"`
	EstimatedTimeMs          int64   `json:"estimated_time_ms"`
	Confidence               float64 `json:"confidence"` // 0.0-1.0
}

// StepData contains information about a specific step
type StepData struct {
	StepNumber  int    `json:"step_number"`
	TotalSteps  int    `json:"total_steps"`
	StepName    string `json:"step_name"`
	Description string `json:"description,omitempty"`
}

// ThinkingData contains information about the thinking phase
type ThinkingData struct {
	Phase       string `json:"phase,omitempty"`       // e.g., "reasoning", "planning", "executing"
	Description string `json:"description,omitempty"` // Human-readable description
}

// ProgressTracker helps track and emit progress events for an agent
type ProgressTracker struct {
	agentID     string
	totalSteps  int
	currentStep int
	stepName    string
	startTime   time.Time
	events      chan<- *AgentEvent
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(agentID string, totalSteps int, events chan<- *AgentEvent) *ProgressTracker {
	return &ProgressTracker{
		agentID:     agentID,
		totalSteps:  totalSteps,
		currentStep: 0,
		startTime:   time.Now(),
		events:      events,
	}
}

// SetTotalSteps updates the total number of steps
func (pt *ProgressTracker) SetTotalSteps(total int) {
	pt.totalSteps = total
}

// StartStep marks the beginning of a new step
func (pt *ProgressTracker) StartStep(stepNumber int, stepName string, description string) {
	pt.currentStep = stepNumber
	pt.stepName = stepName

	pt.emit(AgentEventStepStarted, map[string]interface{}{
		"step_number": stepNumber,
		"total_steps": pt.totalSteps,
		"step_name":   stepName,
		"description": description,
	})

	// Also emit a progress event
	pt.emitProgress("")
}

// CompleteStep marks the current step as complete
func (pt *ProgressTracker) CompleteStep(message string) {
	pt.emit(AgentEventStepCompleted, map[string]interface{}{
		"step_number": pt.currentStep,
		"total_steps": pt.totalSteps,
		"step_name":   pt.stepName,
		"message":     message,
	})
}

// UpdateProgress emits a progress event with an optional message
func (pt *ProgressTracker) UpdateProgress(message string) {
	pt.emitProgress(message)
}

// emitProgress sends a progress event
func (pt *ProgressTracker) emitProgress(message string) {
	percentComplete := 0.0
	if pt.totalSteps > 0 {
		percentComplete = float64(pt.currentStep) / float64(pt.totalSteps) * 100.0
	}

	pt.emit(AgentEventProgress, map[string]interface{}{
		"current_step":     pt.currentStep,
		"total_steps":      pt.totalSteps,
		"percent_complete": percentComplete,
		"step_name":        pt.stepName,
		"message":          message,
	})
}

// StartThinking marks the beginning of a thinking phase
func (pt *ProgressTracker) StartThinking(phase string, description string) {
	pt.emit(AgentEventThinkingStart, map[string]interface{}{
		"phase":       phase,
		"description": description,
	})
}

// EndThinking marks the end of a thinking phase
func (pt *ProgressTracker) EndThinking(phase string) {
	pt.emit(AgentEventThinkingEnd, map[string]interface{}{
		"phase": phase,
	})
}

// EmitEstimate sends an estimate event
func (pt *ProgressTracker) EmitEstimate(tokensRemaining int, timeMs int64, confidence float64) {
	pt.emit(AgentEventEstimate, map[string]interface{}{
		"estimated_tokens_remaining": tokensRemaining,
		"estimated_time_ms":          timeMs,
		"confidence":                 confidence,
	})
}

// emit sends an event to the events channel
func (pt *ProgressTracker) emit(eventType AgentEventType, data map[string]interface{}) {
	if pt.events == nil {
		return
	}

	select {
	case pt.events <- &AgentEvent{
		AgentID:   pt.agentID,
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}:
	default:
		// Channel full, skip
	}
}

// Extended SwarmEventType constants for progress tracking
const (
	// SwarmEventStepProgress indicates progress on a specific step in swarm execution
	SwarmEventStepProgress SwarmEventType = "step_progress"

	// SwarmEventAgentThinking indicates an agent in the swarm is thinking
	SwarmEventAgentThinking SwarmEventType = "agent_thinking"

	// SwarmEventEstimate provides estimated time remaining for swarm completion
	SwarmEventEstimate SwarmEventType = "estimate"
)

// SwarmProgressData contains aggregate swarm progress information
type SwarmProgressData struct {
	CompletedAgents int     `json:"completed_agents"`
	TotalAgents     int     `json:"total_agents"`
	PercentComplete float64 `json:"percent_complete"`
	CurrentPhase    string  `json:"current_phase"` // e.g., "execution", "synthesis"
	Message         string  `json:"message,omitempty"`
}

// SwarmStepProgressData contains step-level progress for swarm strategies
type SwarmStepProgressData struct {
	CurrentStep     int     `json:"current_step"`
	TotalSteps      int     `json:"total_steps"`
	StepName        string  `json:"step_name"`
	PercentComplete float64 `json:"percent_complete"`
}

// SwarmProgressTracker helps track and emit progress events for a swarm
type SwarmProgressTracker struct {
	swarmID        string
	totalAgents    int
	completedCount int
	currentPhase   string
	currentStep    int
	totalSteps     int
	startTime      time.Time
	events         chan<- *SwarmEvent
}

// NewSwarmProgressTracker creates a new swarm progress tracker
func NewSwarmProgressTracker(swarmID string, totalAgents int, events chan<- *SwarmEvent) *SwarmProgressTracker {
	return &SwarmProgressTracker{
		swarmID:     swarmID,
		totalAgents: totalAgents,
		startTime:   time.Now(),
		events:      events,
	}
}

// SetPhase updates the current phase
func (spt *SwarmProgressTracker) SetPhase(phase string) {
	spt.currentPhase = phase
}

// SetTotalSteps sets the total number of steps (for pipeline/sequential strategies)
func (spt *SwarmProgressTracker) SetTotalSteps(total int) {
	spt.totalSteps = total
}

// IncrementCompleted marks one more agent as completed
func (spt *SwarmProgressTracker) IncrementCompleted() {
	spt.completedCount++
	spt.emitProgress("")
}

// StartStep marks the beginning of a step in pipeline strategy
func (spt *SwarmProgressTracker) StartStep(stepNumber int, stepName string) {
	spt.currentStep = stepNumber

	spt.emit(SwarmEventStepProgress, "", "", map[string]interface{}{
		"current_step":     stepNumber,
		"total_steps":      spt.totalSteps,
		"step_name":        stepName,
		"percent_complete": spt.calculatePercent(),
	})
}

// UpdateProgress emits a swarm progress event
func (spt *SwarmProgressTracker) UpdateProgress(message string) {
	spt.emitProgress(message)
}

// emitProgress sends a progress event
func (spt *SwarmProgressTracker) emitProgress(message string) {
	spt.emit(SwarmEventProgress, "", "", map[string]interface{}{
		"completed_agents": spt.completedCount,
		"total_agents":     spt.totalAgents,
		"percent_complete": spt.calculatePercent(),
		"current_phase":    spt.currentPhase,
		"message":          message,
	})
}

// EmitAgentThinking indicates an agent is in thinking phase
func (spt *SwarmProgressTracker) EmitAgentThinking(agentID string, role AgentRole, thinking bool) {
	spt.emit(SwarmEventAgentThinking, agentID, role, map[string]interface{}{
		"thinking": thinking,
	})
}

// EmitEstimate sends a swarm-level estimate
func (spt *SwarmProgressTracker) EmitEstimate(estimatedTimeMs int64, confidence float64) {
	spt.emit(SwarmEventEstimate, "", "", map[string]interface{}{
		"estimated_time_ms": estimatedTimeMs,
		"confidence":        confidence,
	})
}

// calculatePercent calculates the current completion percentage
func (spt *SwarmProgressTracker) calculatePercent() float64 {
	if spt.totalAgents == 0 {
		return 0.0
	}
	return float64(spt.completedCount) / float64(spt.totalAgents) * 100.0
}

// emit sends an event to the swarm events channel
func (spt *SwarmProgressTracker) emit(eventType SwarmEventType, agentID string, role AgentRole, data map[string]interface{}) {
	if spt.events == nil {
		return
	}

	select {
	case spt.events <- &SwarmEvent{
		SwarmID:   spt.swarmID,
		Type:      eventType,
		AgentID:   agentID,
		Role:      role,
		Data:      data,
		Timestamp: time.Now(),
	}:
	default:
		// Channel full, skip
	}
}
