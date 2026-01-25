package cursor

import (
	"time"
)

// API endpoint constants
const (
	DefaultBaseURL = "https://api.cursor.com/v0"
)

// CursorCreateAgentRequest is the request body for creating an agent
type CursorCreateAgentRequest struct {
	Prompt      string `json:"prompt"`
	Model       string `json:"model,omitempty"`
	WorkspaceID string `json:"workspace_id,omitempty"`
}

// CursorAgentResponse is the response from agent creation/retrieval
type CursorAgentResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	Model     string `json:"model,omitempty"`
}

// CursorMessageResponse represents a message from the Cursor API
type CursorMessageResponse struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// CursorMessagesResponse is the response from getting messages
type CursorMessagesResponse struct {
	Messages []CursorMessageResponse `json:"messages"`
}

// CursorFollowupRequest is the request body for sending a follow-up message
type CursorFollowupRequest struct {
	Message string `json:"message"`
}

// CursorErrorResponse represents an error from the Cursor API
type CursorErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// CursorStreamEvent represents an SSE event from Cursor API
type CursorStreamEvent struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`

	// For content events
	Delta string `json:"delta,omitempty"`

	// For tool events
	ToolCall *CursorToolCall `json:"tool_call,omitempty"`

	// For error events
	Error *CursorStreamError `json:"error,omitempty"`

	// For finish events
	FinishReason string `json:"finish_reason,omitempty"`
	MessageID    string `json:"message_id,omitempty"`
}

// CursorToolCall represents a tool call in streaming response
type CursorToolCall struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// CursorStreamError represents an error in streaming response
type CursorStreamError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
