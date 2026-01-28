package agent

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/llm"
)

// ReasoningEffort represents the level of reasoning effort for the orchestrator
type ReasoningEffort string

const (
	ReasoningEffortLow    ReasoningEffort = "low"
	ReasoningEffortMedium ReasoningEffort = "medium"
	ReasoningEffortHigh   ReasoningEffort = "high"
)

// OrchestratorConfig configures an orchestrator workflow
type OrchestratorConfig struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Provider        string          `json:"provider"`
	Model           string          `json:"model"`
	SystemPrompt    string          `json:"system_prompt,omitempty"`
	ReasoningEffort ReasoningEffort `json:"reasoning_effort"`
	MaxThinkingTokens int           `json:"max_thinking_tokens,omitempty"`
	Timeout         time.Duration   `json:"timeout"`
	MaxSubAgents    int             `json:"max_sub_agents"`
	VerboseLogging  bool            `json:"verbose_logging"`
}

// ReasoningLog captures orchestrator decision-making
type ReasoningLog struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Decision   string    `json:"decision"`
	Reasoning  string    `json:"reasoning"`
	SubAgentID string    `json:"sub_agent_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ImageContent represents extracted image content from messages
type ImageContent struct {
	ID       string `json:"id"`
	URL      string `json:"url,omitempty"`
	Base64   string `json:"base64,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	Source   string `json:"source"` // "user" or "assistant"
	MessageIndex int `json:"message_index"`
}

// SubAgentResult represents the result from a spawned sub-agent
type SubAgentResult struct {
	AgentID     string                 `json:"agent_id"`
	Role        AgentRole              `json:"role"`
	Task        string                 `json:"task"`
	Output      string                 `json:"output"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	ToolResults []ToolResult           `json:"tool_results,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// OrchestratorResult represents the final result of an orchestrator workflow
type OrchestratorResult struct {
	ID             string           `json:"id"`
	Success        bool             `json:"success"`
	Output         string           `json:"output"`
	Error          string           `json:"error,omitempty"`
	SubAgentResults []SubAgentResult `json:"sub_agent_results"`
	ReasoningLogs  []ReasoningLog   `json:"reasoning_logs"`
	Images         []ImageContent   `json:"images,omitempty"`
	Duration       time.Duration    `json:"duration"`
	Usage          *llm.Usage       `json:"usage,omitempty"`
	CompletedAt    time.Time        `json:"completed_at"`
}

// OrchestratorWorkflowStatus represents the status of an orchestrator workflow
type OrchestratorWorkflowStatus string

const (
	OrchestratorWorkflowStatusPending   OrchestratorWorkflowStatus = "pending"
	OrchestratorWorkflowStatusRunning   OrchestratorWorkflowStatus = "running"
	OrchestratorWorkflowStatusCompleted OrchestratorWorkflowStatus = "completed"
	OrchestratorWorkflowStatusFailed    OrchestratorWorkflowStatus = "failed"
	OrchestratorWorkflowStatusCancelled OrchestratorWorkflowStatus = "cancelled"
)

// OrchestratorWorkflowEvent represents an event from the orchestrator workflow
type OrchestratorWorkflowEvent struct {
	WorkflowID string                 `json:"workflow_id"`
	Type       OrchestratorEventType  `json:"type"`
	SubAgentID string                 `json:"sub_agent_id,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// OrchestratorEventType represents types of orchestrator events
type OrchestratorEventType string

const (
	OrchestratorEventStarted        OrchestratorEventType = "started"
	OrchestratorEventReasoning      OrchestratorEventType = "reasoning"
	OrchestratorEventSpawningAgent  OrchestratorEventType = "spawning_agent"
	OrchestratorEventAgentCompleted OrchestratorEventType = "agent_completed"
	OrchestratorEventAgentFailed    OrchestratorEventType = "agent_failed"
	OrchestratorEventCompleted      OrchestratorEventType = "completed"
	OrchestratorEventFailed         OrchestratorEventType = "failed"
	OrchestratorEventCancelled      OrchestratorEventType = "cancelled"
)

// OrchestratorWorkflow handles multi-agent coordination with enhanced reasoning
type OrchestratorWorkflow struct {
	ID             string                     `json:"id"`
	Config         OrchestratorConfig         `json:"config"`
	Status         OrchestratorWorkflowStatus `json:"status"`
	CreatedAt      time.Time                  `json:"created_at"`
	StartedAt      *time.Time                 `json:"started_at,omitempty"`
	CompletedAt    *time.Time                 `json:"completed_at,omitempty"`
	Error          string                     `json:"error,omitempty"`

	// Internal state
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	llmManager     *llm.Manager
	agent          *Agent
	subAgents      []*Agent
	reasoningLogs  []ReasoningLog
	extractedImages []ImageContent
	messages       []llm.Message
	results        chan *OrchestratorResult
	events         chan *OrchestratorWorkflowEvent
}

// NewOrchestratorWorkflow creates a new orchestrator workflow with high reasoning config
func NewOrchestratorWorkflow(config OrchestratorConfig, llmManager *llm.Manager) *OrchestratorWorkflow {
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	if config.MaxSubAgents == 0 {
		config.MaxSubAgents = 10
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Minute
	}
	if config.ReasoningEffort == "" {
		config.ReasoningEffort = ReasoningEffortHigh
	}
	if config.MaxThinkingTokens == 0 && config.ReasoningEffort == ReasoningEffortHigh {
		config.MaxThinkingTokens = 16000 // Extended thinking for high reasoning
	}

	// Build orchestrator system prompt
	systemPrompt := config.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = buildOrchestratorSystemPrompt()
	}

	// Create the main orchestrator agent configuration
	agentConfig := AgentConfig{
		ID:           config.ID + "-main",
		Name:         config.Name,
		Provider:     config.Provider,
		Model:        config.Model,
		SystemPrompt: systemPrompt,
		MaxTokens:    config.MaxThinkingTokens,
	}

	return &OrchestratorWorkflow{
		ID:              config.ID,
		Config:          config,
		Status:          OrchestratorWorkflowStatusPending,
		CreatedAt:       time.Now(),
		llmManager:      llmManager,
		agent:           NewAgent(agentConfig, llmManager),
		subAgents:       make([]*Agent, 0),
		reasoningLogs:   make([]ReasoningLog, 0),
		extractedImages: make([]ImageContent, 0),
		messages:        make([]llm.Message, 0),
		results:         make(chan *OrchestratorResult, 1),
		events:          make(chan *OrchestratorWorkflowEvent, 100),
	}
}

// buildOrchestratorSystemPrompt creates the default system prompt for the orchestrator
func buildOrchestratorSystemPrompt() string {
	return `You are an orchestrator agent responsible for coordinating multiple specialized sub-agents to complete complex tasks.

Your responsibilities:
1. **Task Analysis**: Carefully analyze incoming tasks to understand their complexity and requirements
2. **Sub-Agent Coordination**: Decide when and which sub-agents to spawn for specific subtasks
3. **Reasoning Documentation**: Document your reasoning process for transparency
4. **Result Synthesis**: Combine outputs from sub-agents into coherent final results

When spawning sub-agents:
- Use the spawn_sub_agent tool with a clear task prompt and appropriate role
- Available roles: general, planner, coder, reviewer, researcher, writer, analyst, debugger, tester, synthesizer
- Each sub-agent should have a focused, specific task

Your reasoning should be explicit and thorough. Consider:
- What is the main goal?
- What subtasks are needed?
- Which specialists would be best suited for each subtask?
- How should results be combined?

Always provide detailed reasoning logs for your decisions.`
}

// Run executes the orchestration workflow
func (ow *OrchestratorWorkflow) Run(ctx context.Context, task *Task) error {
	ow.mu.Lock()
	if ow.Status == OrchestratorWorkflowStatusRunning {
		ow.mu.Unlock()
		return ErrAgentAlreadyRunning
	}

	ow.ctx, ow.cancel = context.WithTimeout(ctx, ow.Config.Timeout)
	ow.Status = OrchestratorWorkflowStatusRunning
	now := time.Now()
	ow.StartedAt = &now
	ow.mu.Unlock()

	// Emit started event
	ow.emitEvent(OrchestratorEventStarted, "", map[string]interface{}{
		"task_id": task.ID,
		"prompt":  task.Prompt,
	})

	// Extract images from any existing context messages
	if len(ow.messages) > 0 {
		ow.extractedImages = ow.ExtractImages(ow.messages)
	}

	// Run the orchestration asynchronously
	go ow.run(task)

	return nil
}

// run executes the main orchestration loop
func (ow *OrchestratorWorkflow) run(task *Task) {
	startTime := time.Now()

	defer func() {
		if r := recover(); r != nil {
			ow.fail("orchestrator panicked")
		}
		close(ow.events)
	}()

	// Log initial reasoning
	ow.addReasoningLog("Task Analysis", "Analyzing incoming task to determine orchestration strategy", "")

	// Start the main orchestrator agent
	if err := ow.agent.Start(ow.ctx, task); err != nil {
		ow.fail(err.Error())
		ow.sendResult(false, "", err.Error(), startTime)
		return
	}

	// Wait for orchestrator result
	select {
	case result := <-ow.agent.Results():
		ow.mu.Lock()
		now := time.Now()
		ow.CompletedAt = &now
		ow.Status = OrchestratorWorkflowStatusCompleted
		ow.mu.Unlock()

		ow.emitEvent(OrchestratorEventCompleted, "", map[string]interface{}{
			"output": result.Output,
		})

		ow.sendResult(result.Success, result.Output, result.Error, startTime)

	case <-ow.ctx.Done():
		ow.mu.Lock()
		now := time.Now()
		ow.CompletedAt = &now
		ow.Status = OrchestratorWorkflowStatusCancelled
		ow.mu.Unlock()

		ow.emitEvent(OrchestratorEventCancelled, "", nil)
		ow.sendResult(false, "", "workflow cancelled", startTime)
	}
}

// SpawnSubAgent creates and executes a sub-agent
func (ow *OrchestratorWorkflow) SpawnSubAgent(config AgentConfig, task *Task) (*SubAgentResult, error) {
	ow.mu.Lock()
	if len(ow.subAgents) >= ow.Config.MaxSubAgents {
		ow.mu.Unlock()
		return nil, ErrTooManyAgents
	}
	ow.mu.Unlock()

	// Log reasoning for spawning
	ow.addReasoningLog(
		"Spawning Sub-Agent",
		config.SystemPrompt,
		config.ID,
	)

	ow.emitEvent(OrchestratorEventSpawningAgent, config.ID, map[string]interface{}{
		"role":   config.Name,
		"prompt": task.Prompt,
	})

	// Create the sub-agent
	subAgent := NewAgent(config, ow.llmManager)

	ow.mu.Lock()
	ow.subAgents = append(ow.subAgents, subAgent)
	ow.mu.Unlock()

	// Start the sub-agent
	startTime := time.Now()
	if err := subAgent.Start(ow.ctx, task); err != nil {
		ow.emitEvent(OrchestratorEventAgentFailed, config.ID, map[string]interface{}{
			"error": err.Error(),
		})
		return &SubAgentResult{
			AgentID:  config.ID,
			Task:     task.Prompt,
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(startTime),
		}, err
	}

	// Wait for sub-agent result
	select {
	case result := <-subAgent.Results():
		ow.emitEvent(OrchestratorEventAgentCompleted, config.ID, map[string]interface{}{
			"output":  result.Output,
			"success": result.Success,
		})

		return &SubAgentResult{
			AgentID:     config.ID,
			Task:        task.Prompt,
			Output:      result.Output,
			Success:     result.Success,
			Error:       result.Error,
			Duration:    result.Duration,
			ToolResults: result.ToolResults,
		}, nil

	case <-ow.ctx.Done():
		return &SubAgentResult{
			AgentID:  config.ID,
			Task:     task.Prompt,
			Success:  false,
			Error:    "context cancelled",
			Duration: time.Since(startTime),
		}, ow.ctx.Err()
	}
}

// ExtractImages extracts images from message history
func (ow *OrchestratorWorkflow) ExtractImages(messages []llm.Message) []ImageContent {
	images := make([]ImageContent, 0)

	for i, msg := range messages {
		for _, img := range msg.Images {
			imageContent := ImageContent{
				ID:           uuid.New().String(),
				URL:          img.URL,
				Base64:       img.Base64,
				MimeType:     img.MimeType,
				Source:       msg.Role,
				MessageIndex: i,
			}
			images = append(images, imageContent)
		}
	}

	ow.mu.Lock()
	ow.extractedImages = images
	ow.mu.Unlock()

	if ow.Config.VerboseLogging && len(images) > 0 {
		ow.addReasoningLog(
			"Image Extraction",
			"Extracted images from message history for multimodal context",
			"",
		)
	}

	return images
}

// GetReasoningLogs returns all reasoning logs
func (ow *OrchestratorWorkflow) GetReasoningLogs() []ReasoningLog {
	ow.mu.RLock()
	defer ow.mu.RUnlock()
	logs := make([]ReasoningLog, len(ow.reasoningLogs))
	copy(logs, ow.reasoningLogs)
	return logs
}

// GetExtractedImages returns all extracted images
func (ow *OrchestratorWorkflow) GetExtractedImages() []ImageContent {
	ow.mu.RLock()
	defer ow.mu.RUnlock()
	images := make([]ImageContent, len(ow.extractedImages))
	copy(images, ow.extractedImages)
	return images
}

// GetSubAgents returns all spawned sub-agents
func (ow *OrchestratorWorkflow) GetSubAgents() []*Agent {
	ow.mu.RLock()
	defer ow.mu.RUnlock()
	agents := make([]*Agent, len(ow.subAgents))
	copy(agents, ow.subAgents)
	return agents
}

// AddMessage adds a message to the workflow's conversation history
func (ow *OrchestratorWorkflow) AddMessage(msg llm.Message) {
	ow.mu.Lock()
	defer ow.mu.Unlock()
	ow.messages = append(ow.messages, msg)
}

// Results returns the results channel
func (ow *OrchestratorWorkflow) Results() <-chan *OrchestratorResult {
	return ow.results
}

// Events returns the events channel
func (ow *OrchestratorWorkflow) Events() <-chan *OrchestratorWorkflowEvent {
	return ow.events
}

// Stop cancels the workflow execution
func (ow *OrchestratorWorkflow) Stop() {
	ow.mu.Lock()
	defer ow.mu.Unlock()

	if ow.cancel != nil {
		ow.cancel()
	}

	// Stop all sub-agents
	for _, agent := range ow.subAgents {
		agent.Stop()
	}

	ow.Status = OrchestratorWorkflowStatusCancelled
}

// addReasoningLog adds a new reasoning log entry
func (ow *OrchestratorWorkflow) addReasoningLog(decision, reasoning, subAgentID string) {
	log := ReasoningLog{
		ID:         uuid.New().String(),
		Timestamp:  time.Now(),
		Decision:   decision,
		Reasoning:  reasoning,
		SubAgentID: subAgentID,
	}

	ow.mu.Lock()
	ow.reasoningLogs = append(ow.reasoningLogs, log)
	ow.mu.Unlock()

	if ow.Config.VerboseLogging {
		ow.emitEvent(OrchestratorEventReasoning, subAgentID, map[string]interface{}{
			"decision":  decision,
			"reasoning": reasoning,
		})
	}
}

// emitEvent sends an event from the workflow
func (ow *OrchestratorWorkflow) emitEvent(eventType OrchestratorEventType, subAgentID string, data map[string]interface{}) {
	select {
	case ow.events <- &OrchestratorWorkflowEvent{
		WorkflowID: ow.ID,
		Type:       eventType,
		SubAgentID: subAgentID,
		Data:       data,
		Timestamp:  time.Now(),
	}:
	default:
		// Channel full, skip
	}
}

// fail marks the workflow as failed
func (ow *OrchestratorWorkflow) fail(errMsg string) {
	ow.mu.Lock()
	ow.Status = OrchestratorWorkflowStatusFailed
	ow.Error = errMsg
	now := time.Now()
	ow.CompletedAt = &now
	ow.mu.Unlock()

	ow.emitEvent(OrchestratorEventFailed, "", map[string]interface{}{
		"error": errMsg,
	})
}

// sendResult sends the final result
func (ow *OrchestratorWorkflow) sendResult(success bool, output, errMsg string, startTime time.Time) {
	ow.mu.RLock()
	subAgentResults := make([]SubAgentResult, 0)
	for _, agent := range ow.subAgents {
		subAgentResults = append(subAgentResults, SubAgentResult{
			AgentID: agent.ID,
			Success: agent.Status == AgentStatusCompleted,
		})
	}
	reasoningLogs := make([]ReasoningLog, len(ow.reasoningLogs))
	copy(reasoningLogs, ow.reasoningLogs)
	images := make([]ImageContent, len(ow.extractedImages))
	copy(images, ow.extractedImages)
	ow.mu.RUnlock()

	ow.results <- &OrchestratorResult{
		ID:              ow.ID,
		Success:         success,
		Output:          output,
		Error:           errMsg,
		SubAgentResults: subAgentResults,
		ReasoningLogs:   reasoningLogs,
		Images:          images,
		Duration:        time.Since(startTime),
		CompletedAt:     time.Now(),
	}
}
