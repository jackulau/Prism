package workflow

import (
	"encoding/json"
	"time"
)

// WorkflowStatus represents the current status of a workflow
type WorkflowStatus string

const (
	StatusPending   WorkflowStatus = "pending"
	StatusRunning   WorkflowStatus = "running"
	StatusPaused    WorkflowStatus = "paused"
	StatusCompleted WorkflowStatus = "completed"
	StatusFailed    WorkflowStatus = "failed"
	StatusCancelled WorkflowStatus = "cancelled"
)

// StepStatus represents the status of a workflow step
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// StepType represents the type of a workflow step
type StepType string

const (
	StepTypeAgent     StepType = "agent"     // Run agent with prompt
	StepTypeTool      StepType = "tool"      // Execute specific tool
	StepTypeCondition StepType = "condition" // Evaluate condition
	StepTypeParallel  StepType = "parallel"  // Run multiple steps in parallel
	StepTypeWait      StepType = "wait"      // Wait for external input
	StepTypeTransform StepType = "transform" // Transform data
)

// Workflow represents a workflow definition and its execution state
type Workflow struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Steps       []Step                 `json:"steps"`
	Status      WorkflowStatus         `json:"status"`
	CurrentStep int                    `json:"current_step"`
	State       map[string]interface{} `json:"state,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// Step represents a single step in a workflow
type Step struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Type        StepType      `json:"type"`
	Config      StepConfig    `json:"config"`
	Condition   *Condition    `json:"condition,omitempty"`   // Optional: skip if condition false
	OnSuccess   string        `json:"on_success,omitempty"`  // Next step ID on success
	OnFailure   string        `json:"on_failure,omitempty"`  // Next step ID on failure
	Timeout     time.Duration `json:"timeout,omitempty"`
	RetryPolicy *RetryPolicy  `json:"retry_policy,omitempty"`
}

// StepConfig holds configuration for different step types
type StepConfig struct {
	AgentConfig     *AgentStepConfig     `json:"agent_config,omitempty"`
	ToolConfig      *ToolStepConfig      `json:"tool_config,omitempty"`
	ConditionConfig *ConditionConfig     `json:"condition_config,omitempty"`
	ParallelConfig  *ParallelStepConfig  `json:"parallel_config,omitempty"`
	WaitConfig      *WaitStepConfig      `json:"wait_config,omitempty"`
	TransformConfig *TransformStepConfig `json:"transform_config,omitempty"`
}

// AgentStepConfig configures an agent execution step
type AgentStepConfig struct {
	Provider     string   `json:"provider"`
	Model        string   `json:"model"`
	SystemPrompt string   `json:"system_prompt,omitempty"`
	Prompt       string   `json:"prompt"`                 // Can include {{state.variable}} placeholders
	Temperature  float64  `json:"temperature,omitempty"`
	MaxTokens    int      `json:"max_tokens,omitempty"`
	Tools        []string `json:"tools,omitempty"`        // Tool names to enable
	OutputKey    string   `json:"output_key,omitempty"`   // Key to store output in state
}

// ToolStepConfig configures a tool execution step
type ToolStepConfig struct {
	ToolName   string                 `json:"tool_name"`
	Parameters map[string]interface{} `json:"parameters"`         // Can include {{state.variable}} placeholders
	OutputKey  string                 `json:"output_key,omitempty"` // Key to store output in state
}

// ConditionConfig configures a condition evaluation
type ConditionConfig struct {
	Expression  string `json:"expression"`   // Expression to evaluate
	TrueBranch  string `json:"true_branch"`  // Step ID if true
	FalseBranch string `json:"false_branch"` // Step ID if false
}

// ParallelStepConfig configures parallel step execution
type ParallelStepConfig struct {
	Steps       []Step `json:"steps"`         // Steps to execute in parallel
	WaitForAll  bool   `json:"wait_for_all"`  // Wait for all to complete
	FailOnFirst bool   `json:"fail_on_first"` // Fail immediately on first error
}

// WaitStepConfig configures a wait step
type WaitStepConfig struct {
	WaitType    string        `json:"wait_type"`            // "user_input", "webhook", "timeout"
	Timeout     time.Duration `json:"timeout,omitempty"`
	PromptText  string        `json:"prompt_text,omitempty"` // For user input
	WebhookPath string        `json:"webhook_path,omitempty"` // For webhook wait
	OutputKey   string        `json:"output_key,omitempty"`
}

// TransformStepConfig configures a data transformation step
type TransformStepConfig struct {
	Type      string                 `json:"type"`       // "jq", "template", "script"
	Template  string                 `json:"template,omitempty"` // Template string
	Script    string                 `json:"script,omitempty"`   // Script code
	InputKey  string                 `json:"input_key,omitempty"`
	OutputKey string                 `json:"output_key,omitempty"`
	Mapping   map[string]string      `json:"mapping,omitempty"` // Simple key mapping
}

// Condition represents a condition for step execution
type Condition struct {
	Type       string `json:"type"`                 // "expression", "state_check"
	Expression string `json:"expression,omitempty"` // Expression to evaluate
	StateKey   string `json:"state_key,omitempty"`  // Key to check in state
	Operator   string `json:"operator,omitempty"`   // "equals", "not_equals", "exists", "contains"
	Value      string `json:"value,omitempty"`      // Value to compare against
}

// RetryPolicy defines retry behavior for a step
type RetryPolicy struct {
	MaxRetries  int           `json:"max_retries"`
	Delay       time.Duration `json:"delay"`
	BackoffType string        `json:"backoff_type,omitempty"` // "fixed", "exponential"
	MaxDelay    time.Duration `json:"max_delay,omitempty"`
}

// StepResult represents the result of a step execution
type StepResult struct {
	StepID      string                 `json:"step_id"`
	StepName    string                 `json:"step_name"`
	Status      StepStatus             `json:"status"`
	Output      interface{}            `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	RetryCount  int                    `json:"retry_count,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowDefinition is used for creating new workflows
type WorkflowDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Steps       []Step                 `json:"steps"`
	InitialState map[string]interface{} `json:"initial_state,omitempty"`
}

// WorkflowFilter for listing workflows
type WorkflowFilter struct {
	UserID   string           `json:"user_id,omitempty"`
	Status   []WorkflowStatus `json:"status,omitempty"`
	Name     string           `json:"name,omitempty"`
	Limit    int              `json:"limit,omitempty"`
	Offset   int              `json:"offset,omitempty"`
}

// WorkflowEvent represents an event from a workflow during execution
type WorkflowEvent struct {
	WorkflowID string                 `json:"workflow_id"`
	Type       WorkflowEventType      `json:"type"`
	StepID     string                 `json:"step_id,omitempty"`
	StepName   string                 `json:"step_name,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// WorkflowEventType represents types of workflow events
type WorkflowEventType string

const (
	WorkflowEventStarted       WorkflowEventType = "workflow_started"
	WorkflowEventPaused        WorkflowEventType = "workflow_paused"
	WorkflowEventResumed       WorkflowEventType = "workflow_resumed"
	WorkflowEventCompleted     WorkflowEventType = "workflow_completed"
	WorkflowEventFailed        WorkflowEventType = "workflow_failed"
	WorkflowEventCancelled     WorkflowEventType = "workflow_cancelled"
	WorkflowEventStepStarted   WorkflowEventType = "step_started"
	WorkflowEventStepCompleted WorkflowEventType = "step_completed"
	WorkflowEventStepFailed    WorkflowEventType = "step_failed"
	WorkflowEventStepSkipped   WorkflowEventType = "step_skipped"
	WorkflowEventStepRetrying  WorkflowEventType = "step_retrying"
	WorkflowEventStateUpdated  WorkflowEventType = "state_updated"
	WorkflowEventWaitingInput  WorkflowEventType = "waiting_input"
)

// MarshalJSON implements custom JSON marshaling for Duration fields
func (s Step) MarshalJSON() ([]byte, error) {
	type Alias Step
	return json.Marshal(&struct {
		Timeout int64 `json:"timeout,omitempty"`
		*Alias
	}{
		Timeout: int64(s.Timeout / time.Millisecond),
		Alias:   (*Alias)(&s),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Duration fields
func (s *Step) UnmarshalJSON(data []byte) error {
	type Alias Step
	aux := &struct {
		Timeout int64 `json:"timeout,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	s.Timeout = time.Duration(aux.Timeout) * time.Millisecond
	return nil
}

// Clone creates a deep copy of the workflow
func (w *Workflow) Clone() *Workflow {
	if w == nil {
		return nil
	}

	clone := *w

	// Deep copy steps
	clone.Steps = make([]Step, len(w.Steps))
	copy(clone.Steps, w.Steps)

	// Deep copy state
	if w.State != nil {
		clone.State = make(map[string]interface{})
		for k, v := range w.State {
			clone.State[k] = v
		}
	}

	return &clone
}

// GetCurrentStep returns the current step being executed
func (w *Workflow) GetCurrentStep() *Step {
	if w.CurrentStep < 0 || w.CurrentStep >= len(w.Steps) {
		return nil
	}
	return &w.Steps[w.CurrentStep]
}

// GetStepByID finds a step by its ID
func (w *Workflow) GetStepByID(stepID string) *Step {
	for i := range w.Steps {
		if w.Steps[i].ID == stepID {
			return &w.Steps[i]
		}
	}
	return nil
}

// GetStepIndex returns the index of a step by its ID
func (w *Workflow) GetStepIndex(stepID string) int {
	for i, step := range w.Steps {
		if step.ID == stepID {
			return i
		}
	}
	return -1
}

// SetStateValue sets a value in the workflow state
func (w *Workflow) SetStateValue(key string, value interface{}) {
	if w.State == nil {
		w.State = make(map[string]interface{})
	}
	w.State[key] = value
}

// GetStateValue gets a value from the workflow state
func (w *Workflow) GetStateValue(key string) (interface{}, bool) {
	if w.State == nil {
		return nil, false
	}
	val, ok := w.State[key]
	return val, ok
}

// IsTerminal returns true if the workflow is in a terminal state
func (w *Workflow) IsTerminal() bool {
	return w.Status == StatusCompleted || w.Status == StatusFailed || w.Status == StatusCancelled
}

// CanStart returns true if the workflow can be started
func (w *Workflow) CanStart() bool {
	return w.Status == StatusPending
}

// CanPause returns true if the workflow can be paused
func (w *Workflow) CanPause() bool {
	return w.Status == StatusRunning
}

// CanResume returns true if the workflow can be resumed
func (w *Workflow) CanResume() bool {
	return w.Status == StatusPaused
}

// CanCancel returns true if the workflow can be cancelled
func (w *Workflow) CanCancel() bool {
	return w.Status == StatusPending || w.Status == StatusRunning || w.Status == StatusPaused
}
