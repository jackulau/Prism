package sse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of SSE event
type EventType string

// Event types matching WebSocket message types
const (
	// Connection events
	EventConnected EventType = "connected"

	// Chat events
	EventChatChunk    EventType = "chat.chunk"
	EventChatComplete EventType = "chat.complete"

	// Tool events
	EventToolStarted   EventType = "tool.started"
	EventToolCompleted EventType = "tool.completed"
	EventToolConfirm   EventType = "tool.confirm"

	// Error event
	EventError EventType = "error"

	// Heartbeat for keeping connection alive
	EventHeartbeat EventType = "heartbeat"

	// Agent events
	EventAgentStarted       EventType = "agent.started"
	EventAgentThinking      EventType = "agent.thinking"
	EventAgentStreamChunk   EventType = "agent.stream_chunk"
	EventAgentToolCall      EventType = "agent.tool_call"
	EventAgentToolResult    EventType = "agent.tool_result"
	EventAgentCompleted     EventType = "agent.completed"
	EventAgentFailed        EventType = "agent.failed"
	EventAgentCancelled     EventType = "agent.cancelled"
	EventAgentStatus        EventType = "agent.status"
	EventAgentList          EventType = "agent.list"
	EventAgentBatchProgress EventType = "agent.batch_progress"
	EventAgentBatchComplete EventType = "agent.batch_completed"
	EventAgentCheckIn       EventType = "agent.check_in"

	// Preview/Sandbox events
	EventPreviewReady   EventType = "preview.ready"
	EventPreviewContent EventType = "preview.content"
	EventPreviewError   EventType = "preview.error"
	EventBuildStarted   EventType = "build.started"
	EventBuildOutput    EventType = "build.output"
	EventBuildCompleted EventType = "build.completed"

	// Shell events
	EventShellOutput    EventType = "shell.output"
	EventShellCompleted EventType = "shell.completed"
	EventShellFailed    EventType = "shell.failed"

	// File events
	EventFilesUpdated      EventType = "files.updated"
	EventFileContent       EventType = "file.content"
	EventFileHistoryList   EventType = "file.history_list"
	EventFileHistoryContent EventType = "file.history_content"

	// Swarm events
	EventSwarmStarted        EventType = "swarm.started"
	EventSwarmAgentStarted   EventType = "swarm.agent_started"
	EventSwarmAgentOutput    EventType = "swarm.agent_output"
	EventSwarmAgentCompleted EventType = "swarm.agent_completed"
	EventSwarmAgentFailed    EventType = "swarm.agent_failed"
	EventSwarmProgress       EventType = "swarm.progress"
	EventSwarmSynthesizing   EventType = "swarm.synthesizing"
	EventSwarmCompleted      EventType = "swarm.completed"
	EventSwarmFailed         EventType = "swarm.failed"
	EventSwarmCancelled      EventType = "swarm.cancelled"
	EventSwarmStatus         EventType = "swarm.status"
	EventSwarmList           EventType = "swarm.list"

	// Approval events
	EventApprovalRequested   EventType = "approval.requested"
	EventApprovalApproved    EventType = "approval.approved"
	EventApprovalRejected    EventType = "approval.rejected"
	EventApprovalEscalated   EventType = "approval.escalated"
	EventApprovalExpired     EventType = "approval.expired"
	EventApprovalCancelled   EventType = "approval.cancelled"
	EventApprovalStepUpdated EventType = "approval.step_updated"
	EventApprovalReminder    EventType = "approval.reminder"
)

// Event represents an SSE event
type Event struct {
	ID    string      `json:"id,omitempty"`
	Type  EventType   `json:"event"`
	Data  interface{} `json:"data"`
	Retry int         `json:"retry,omitempty"` // Reconnection time in milliseconds
}

// NewEvent creates a new SSE event with the given type and data
func NewEvent(eventType EventType, data interface{}) *Event {
	return &Event{
		ID:   uuid.New().String(),
		Type: eventType,
		Data: data,
	}
}

// NewEventWithRetry creates a new SSE event with a retry hint
func NewEventWithRetry(eventType EventType, data interface{}, retryMs int) *Event {
	return &Event{
		ID:    uuid.New().String(),
		Type:  eventType,
		Data:  data,
		Retry: retryMs,
	}
}

// NewHeartbeat creates a heartbeat event
func NewHeartbeat() *Event {
	return &Event{
		Type: EventHeartbeat,
		Data: map[string]interface{}{
			"timestamp": time.Now().UnixMilli(),
		},
	}
}

// Format formats the event as an SSE message string
// SSE format: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
func (e *Event) Format() string {
	var buf bytes.Buffer

	// Write event ID if present
	if e.ID != "" {
		fmt.Fprintf(&buf, "id: %s\n", e.ID)
	}

	// Write retry if present
	if e.Retry > 0 {
		fmt.Fprintf(&buf, "retry: %d\n", e.Retry)
	}

	// Write event type
	fmt.Fprintf(&buf, "event: %s\n", e.Type)

	// Write data
	data, err := json.Marshal(e.Data)
	if err != nil {
		data = []byte("{}")
	}
	fmt.Fprintf(&buf, "data: %s\n", data)

	// End with double newline to complete the event
	buf.WriteString("\n")

	return buf.String()
}

// ChatChunkData represents data for a chat chunk event
type ChatChunkData struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	Delta          string `json:"delta"`
}

// ChatCompleteData represents data for a chat complete event
type ChatCompleteData struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	FinishReason   string `json:"finish_reason"`
}

// ToolStartedData represents data for a tool started event
type ToolStartedData struct {
	ConversationID string      `json:"conversation_id"`
	ExecutionID    string      `json:"execution_id"`
	ToolName       string      `json:"tool_name"`
	Parameters     interface{} `json:"parameters"`
	IsMCPTool      bool        `json:"is_mcp_tool,omitempty"`
	MCPServerName  string      `json:"mcp_server_name,omitempty"`
	IsStdioMCP     bool        `json:"is_stdio_mcp,omitempty"`
}

// ToolCompletedData represents data for a tool completed event
type ToolCompletedData struct {
	ConversationID string      `json:"conversation_id"`
	ExecutionID    string      `json:"execution_id"`
	Result         interface{} `json:"result"`
	Status         string      `json:"status"`
}

// ToolConfirmData represents data for a tool confirmation event
type ToolConfirmData struct {
	ConversationID string      `json:"conversation_id"`
	ExecutionID    string      `json:"execution_id"`
	ToolName       string      `json:"tool_name"`
	Parameters     interface{} `json:"parameters"`
	IsMCPTool      bool        `json:"is_mcp_tool,omitempty"`
	MCPServerName  string      `json:"mcp_server_name,omitempty"`
	IsStdioMCP     bool        `json:"is_stdio_mcp,omitempty"`
	IterationCount int         `json:"iteration_count,omitempty"`
}

// ErrorData represents data for an error event
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Helper functions to create common events

// NewChatChunk creates a chat chunk event
func NewChatChunk(conversationID, messageID, delta string) *Event {
	return NewEvent(EventChatChunk, ChatChunkData{
		ConversationID: conversationID,
		MessageID:      messageID,
		Delta:          delta,
	})
}

// NewChatComplete creates a chat complete event
func NewChatComplete(conversationID, messageID, finishReason string) *Event {
	return NewEvent(EventChatComplete, ChatCompleteData{
		ConversationID: conversationID,
		MessageID:      messageID,
		FinishReason:   finishReason,
	})
}

// NewToolStarted creates a tool started event
func NewToolStarted(conversationID, executionID, toolName string, parameters interface{}) *Event {
	return NewEvent(EventToolStarted, ToolStartedData{
		ConversationID: conversationID,
		ExecutionID:    executionID,
		ToolName:       toolName,
		Parameters:     parameters,
	})
}

// NewToolCompleted creates a tool completed event
func NewToolCompleted(conversationID, executionID string, result interface{}, status string) *Event {
	return NewEvent(EventToolCompleted, ToolCompletedData{
		ConversationID: conversationID,
		ExecutionID:    executionID,
		Result:         result,
		Status:         status,
	})
}

// NewToolConfirm creates a tool confirmation request event
func NewToolConfirm(conversationID, executionID, toolName string, parameters interface{}) *Event {
	return NewEvent(EventToolConfirm, ToolConfirmData{
		ConversationID: conversationID,
		ExecutionID:    executionID,
		ToolName:       toolName,
		Parameters:     parameters,
	})
}

// NewError creates an error event
func NewError(code, message string) *Event {
	return NewEvent(EventError, ErrorData{
		Code:    code,
		Message: message,
	})
}

// ApprovalRequestData represents data for approval request events
type ApprovalRequestData struct {
	RequestID        string                 `json:"request_id"`
	WorkflowID       string                 `json:"workflow_id"`
	WorkflowName     string                 `json:"workflow_name,omitempty"`
	OrganizationID   string                 `json:"organization_id"`
	RequesterID      string                 `json:"requester_id"`
	RequesterEmail   string                 `json:"requester_email,omitempty"`
	OperationType    string                 `json:"operation_type"`
	OperationDetails map[string]interface{} `json:"operation_details,omitempty"`
	CurrentStep      int                    `json:"current_step"`
	TotalSteps       int                    `json:"total_steps"`
	Status           string                 `json:"status"`
	Priority         int                    `json:"priority,omitempty"`
	ExpiresAt        string                 `json:"expires_at,omitempty"`
}

// ApprovalDecisionData represents data for approval decision events
type ApprovalDecisionData struct {
	RequestID     string `json:"request_id"`
	ApproverID    string `json:"approver_id"`
	ApproverEmail string `json:"approver_email,omitempty"`
	Decision      string `json:"decision"`
	Comment       string `json:"comment,omitempty"`
	StepOrder     int    `json:"step_order"`
}

// ApprovalStepData represents data for approval step update events
type ApprovalStepData struct {
	RequestID     string `json:"request_id"`
	StepOrder     int    `json:"step_order"`
	StepName      string `json:"step_name"`
	ApprovedCount int    `json:"approved_count"`
	RequiredCount int    `json:"required_count"`
	Status        string `json:"status"`
}

// NewApprovalRequested creates an approval requested event
func NewApprovalRequested(data *ApprovalRequestData) *Event {
	return NewEvent(EventApprovalRequested, data)
}

// NewApprovalApproved creates an approval approved event
func NewApprovalApproved(requestID, approverID, comment string) *Event {
	return NewEvent(EventApprovalApproved, ApprovalDecisionData{
		RequestID:  requestID,
		ApproverID: approverID,
		Decision:   "approved",
		Comment:    comment,
	})
}

// NewApprovalRejected creates an approval rejected event
func NewApprovalRejected(requestID, approverID, comment string) *Event {
	return NewEvent(EventApprovalRejected, ApprovalDecisionData{
		RequestID:  requestID,
		ApproverID: approverID,
		Decision:   "rejected",
		Comment:    comment,
	})
}

// NewApprovalEscalated creates an approval escalated event
func NewApprovalEscalated(requestID string, newPriority int) *Event {
	return NewEvent(EventApprovalEscalated, map[string]interface{}{
		"request_id":   requestID,
		"new_priority": newPriority,
	})
}

// NewApprovalExpired creates an approval expired event
func NewApprovalExpired(requestID string) *Event {
	return NewEvent(EventApprovalExpired, map[string]interface{}{
		"request_id": requestID,
	})
}

// NewApprovalCancelled creates an approval cancelled event
func NewApprovalCancelled(requestID, cancelledBy string) *Event {
	return NewEvent(EventApprovalCancelled, map[string]interface{}{
		"request_id":   requestID,
		"cancelled_by": cancelledBy,
	})
}

// NewApprovalStepUpdated creates an approval step update event
func NewApprovalStepUpdated(data *ApprovalStepData) *Event {
	return NewEvent(EventApprovalStepUpdated, data)
}

// NewApprovalReminder creates an approval reminder event
func NewApprovalReminder(requestID, stepName string, expiresAt string) *Event {
	return NewEvent(EventApprovalReminder, map[string]interface{}{
		"request_id": requestID,
		"step_name":  stepName,
		"expires_at": expiresAt,
	})
}
