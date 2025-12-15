package tools

import (
	"strings"
)

// AutoApprovalConfig defines settings for automatic tool execution
type AutoApprovalConfig struct {
	// AutoApproveReadOnly auto-approves tools that are read-only (don't modify state)
	AutoApproveReadOnly bool `json:"auto_approve_read_only"`

	// TrustedTools is a list of specific tool names that can be auto-approved
	TrustedTools []string `json:"trusted_tools"`

	// MaxIterations is the maximum number of tool calls before requiring user check-in
	// Default is 10 if not set
	MaxIterations int `json:"max_iterations"`

	// Enabled controls whether auto-approval is active at all
	Enabled bool `json:"enabled"`
}

// DefaultAutoApprovalConfig returns the default configuration
func DefaultAutoApprovalConfig() *AutoApprovalConfig {
	return &AutoApprovalConfig{
		AutoApproveReadOnly: false,
		TrustedTools:        []string{},
		MaxIterations:       10,
		Enabled:             false,
	}
}

// ReadOnlyTools is a list of built-in tools that are considered read-only
// These don't modify any state and are safe to auto-approve
var ReadOnlyTools = map[string]bool{
	"read_file":    true,
	"list_files":   true,
	"search_code":  true,
	"grep":         true,
	"glob":         true,
	"get_info":     true,
	"web_fetch":    true,
	"web_search":   true,
	"database_query": true, // SELECT queries only
}

// ShouldAutoApprove determines if a tool should be automatically approved
func (c *AutoApprovalConfig) ShouldAutoApprove(toolName string, isMCP bool) bool {
	// If auto-approval is not enabled, always require confirmation
	if !c.Enabled {
		return false
	}

	// Check if tool is in the trusted tools list
	for _, trusted := range c.TrustedTools {
		// Support wildcards for MCP tools (e.g., "mcp_filesystem_*")
		if strings.HasSuffix(trusted, "*") {
			prefix := strings.TrimSuffix(trusted, "*")
			if strings.HasPrefix(toolName, prefix) {
				return true
			}
		}
		if trusted == toolName {
			return true
		}
	}

	// Check if it's a read-only tool and auto-approve for read-only is enabled
	if c.AutoApproveReadOnly {
		// For built-in tools, check the read-only list
		if !isMCP {
			if ReadOnlyTools[toolName] {
				return true
			}
		}
		// For MCP tools, we currently can't determine read-only status
		// Future: could add metadata to MCP tools to indicate this
	}

	return false
}

// ShouldCheckIn returns true if the agent should pause for user check-in
// based on the iteration count
func (c *AutoApprovalConfig) ShouldCheckIn(iterationCount int) bool {
	if c.MaxIterations <= 0 {
		return false // No limit
	}
	return iterationCount >= c.MaxIterations
}

// IterationState tracks the current state of an agentic loop
type IterationState struct {
	Count          int    `json:"count"`
	ConversationID string `json:"conversation_id"`
	LastToolCallID string `json:"last_tool_call_id"`
}

// NewIterationState creates a new iteration state
func NewIterationState(conversationID string) *IterationState {
	return &IterationState{
		Count:          0,
		ConversationID: conversationID,
	}
}

// Increment increases the iteration count and returns the new count
func (s *IterationState) Increment() int {
	s.Count++
	return s.Count
}

// Reset resets the iteration count to 0
func (s *IterationState) Reset() {
	s.Count = 0
}
