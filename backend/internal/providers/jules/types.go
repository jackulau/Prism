package jules

import (
	"time"
)

// API endpoint constants
const (
	DefaultBaseURL = "https://api.jules.ai/v1"
)

// JulesCreateAgentRequest is the request body for creating an agent
type JulesCreateAgentRequest struct {
	Prompt     string            `json:"prompt"`
	Model      string            `json:"model,omitempty"`
	Repository string            `json:"repository,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// JulesAgentResponse is the response from agent creation/retrieval
type JulesAgentResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Branch    string `json:"branch,omitempty"`
	CreatedAt string `json:"created_at"`
	Model     string `json:"model,omitempty"`
}

// JulesMessageResponse represents a message from the Jules API
type JulesMessageResponse struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp string    `json:"timestamp"`
	CreatedAt time.Time `json:"-"`
}

// JulesMessagesResponse is the response from getting messages
type JulesMessagesResponse struct {
	Messages []JulesMessageResponse `json:"messages"`
}

// JulesFollowupRequest is the request body for sending a follow-up message
type JulesFollowupRequest struct {
	Message string `json:"message"`
}

// JulesErrorResponse represents an error from the Jules API
type JulesErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// JulesStreamEvent represents an SSE event from Jules API
type JulesStreamEvent struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`

	// For content events
	Delta string `json:"delta,omitempty"`

	// For tool events
	ToolCall *JulesToolCall `json:"tool_call,omitempty"`

	// For error events
	Error *JulesStreamError `json:"error,omitempty"`

	// For finish events
	FinishReason string `json:"finish_reason,omitempty"`
	MessageID    string `json:"message_id,omitempty"`
}

// JulesToolCall represents a tool call in streaming response
type JulesToolCall struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// JulesStreamError represents an error in streaming response
type JulesStreamError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
