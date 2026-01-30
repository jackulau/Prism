package types

import "time"

// AttributionContext captures who/what made a change to a file
type AttributionContext struct {
	AgentID        string            `json:"agent_id,omitempty"`
	AgentName      string            `json:"agent_name,omitempty"`
	AgentType      string            `json:"agent_type,omitempty"` // "assistant", "autonomous", "workflow"
	ToolName       string            `json:"tool_name,omitempty"`
	ToolSlug       string            `json:"tool_slug,omitempty"`
	MessageID      string            `json:"message_id,omitempty"`
	ConversationID string            `json:"conversation_id,omitempty"`
	WorkflowID     string            `json:"workflow_id,omitempty"`
	StepID         string            `json:"step_id,omitempty"`
	Description    string            `json:"description,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	Timestamp      time.Time         `json:"timestamp"`
}

// AttributionSummary aggregates attribution data for reporting
type AttributionSummary struct {
	TotalChanges    int            `json:"total_changes"`
	ByAgent         map[string]int `json:"by_agent"`
	ByTool          map[string]int `json:"by_tool"`
	ByOperation     map[string]int `json:"by_operation"`
	TimelineByDay   map[string]int `json:"timeline_by_day"`
	MostActiveAgent string         `json:"most_active_agent"`
	MostUsedTool    string         `json:"most_used_tool"`
}

// AgentActivityReport provides detailed activity information for a specific agent
type AgentActivityReport struct {
	AgentID        string               `json:"agent_id"`
	AgentName      string               `json:"agent_name"`
	TotalChanges   int                  `json:"total_changes"`
	FileChanges    []FileChangeInfo     `json:"file_changes"`
	OperationStats map[string]int       `json:"operation_stats"`
	ActivePeriod   *ActivePeriod        `json:"active_period,omitempty"`
	TopFiles       []FileActivityInfo   `json:"top_files"`
}

// FileChangeInfo represents a single file change made by an agent
type FileChangeInfo struct {
	FilePath      string    `json:"file_path"`
	Operation     string    `json:"operation"`
	ToolName      string    `json:"tool_name,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	ConversationID string   `json:"conversation_id,omitempty"`
}

// ActivePeriod represents the time range when an agent was active
type ActivePeriod struct {
	FirstChange time.Time `json:"first_change"`
	LastChange  time.Time `json:"last_change"`
}

// FileActivityInfo represents aggregated activity for a specific file
type FileActivityInfo struct {
	FilePath    string `json:"file_path"`
	ChangeCount int    `json:"change_count"`
}

// NewAttributionContext creates a new AttributionContext with the current timestamp
func NewAttributionContext() *AttributionContext {
	return &AttributionContext{
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}
}

// WithAgent sets agent information on the attribution context
func (a *AttributionContext) WithAgent(agentID, agentName, agentType string) *AttributionContext {
	a.AgentID = agentID
	a.AgentName = agentName
	a.AgentType = agentType
	return a
}

// WithTool sets tool information on the attribution context
func (a *AttributionContext) WithTool(toolName, toolSlug string) *AttributionContext {
	a.ToolName = toolName
	a.ToolSlug = toolSlug
	return a
}

// WithMessage sets message/conversation information on the attribution context
func (a *AttributionContext) WithMessage(messageID, conversationID string) *AttributionContext {
	a.MessageID = messageID
	a.ConversationID = conversationID
	return a
}

// WithWorkflow sets workflow information on the attribution context
func (a *AttributionContext) WithWorkflow(workflowID, stepID string) *AttributionContext {
	a.WorkflowID = workflowID
	a.StepID = stepID
	return a
}

// WithDescription sets a description on the attribution context
func (a *AttributionContext) WithDescription(description string) *AttributionContext {
	a.Description = description
	return a
}

// SetMetadata sets a metadata key-value pair
func (a *AttributionContext) SetMetadata(key, value string) *AttributionContext {
	if a.Metadata == nil {
		a.Metadata = make(map[string]string)
	}
	a.Metadata[key] = value
	return a
}

// NewAttributionSummary creates a new empty AttributionSummary
func NewAttributionSummary() *AttributionSummary {
	return &AttributionSummary{
		ByAgent:       make(map[string]int),
		ByTool:        make(map[string]int),
		ByOperation:   make(map[string]int),
		TimelineByDay: make(map[string]int),
	}
}
