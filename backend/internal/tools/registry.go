package tools

import (
	"context"
	"fmt"
	"sync"

	"github.com/jacklau/prism/internal/llm"
)

// Tool defines the interface for all tools that can be called by the LLM
type Tool interface {
	// Name returns the unique identifier for this tool
	Name() string

	// Description returns a human-readable description of what this tool does
	Description() string

	// Parameters returns the JSON Schema for the tool's parameters
	Parameters() llm.JSONSchema

	// Execute runs the tool with the given parameters
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)

	// RequiresConfirmation returns true if this tool needs user approval before execution
	RequiresConfirmation() bool
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PendingExecution represents a tool execution awaiting user confirmation
type PendingExecution struct {
	ID             string                 `json:"id"`
	ToolCallID     string                 `json:"tool_call_id"`     // Original tool call ID from the LLM
	ToolName       string                 `json:"tool_name"`
	Parameters     map[string]interface{} `json:"parameters"`
	ConversationID string                 `json:"conversation_id"`
	MessageID      string                 `json:"message_id"`
	UserID         string                 `json:"user_id"`
	// MCP-specific fields
	IsMCPTool   bool   `json:"is_mcp_tool,omitempty"`
	MCPServerID string `json:"mcp_server_id,omitempty"`
	MCPToolName string `json:"mcp_tool_name,omitempty"`
}

// ExecutionResult represents the result of a tool execution
type ExecutionResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Registry manages the collection of available tools
type Registry struct {
	tools             map[string]Tool
	pendingExecutions map[string]*PendingExecution
	mu                sync.RWMutex
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools:             make(map[string]Tool),
		pendingExecutions: make(map[string]*PendingExecution),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[tool.Name()]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name())
	}

	r.tools[tool.Name()] = tool
	return nil
}

// Get returns a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ToLLMTools converts all registered tools to LLM tool definitions
func (r *Registry) ToLLMTools() []llm.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]llm.ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		defs = append(defs, llm.ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  tool.Parameters(),
		})
	}
	return defs
}

// Execute runs a tool by name with the given parameters
func (r *Registry) Execute(ctx context.Context, name string, params map[string]interface{}) (*ToolResult, error) {
	tool, ok := r.Get(name)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("tool %s not found", name),
		}, nil
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &ToolResult{
		Success: true,
		Result:  result,
	}, nil
}

// AddPendingExecution stores a pending execution for later confirmation
func (r *Registry) AddPendingExecution(exec *PendingExecution) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pendingExecutions[exec.ID] = exec
}

// GetPendingExecution retrieves a pending execution by ID
func (r *Registry) GetPendingExecution(id string) (*PendingExecution, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	exec, ok := r.pendingExecutions[id]
	return exec, ok
}

// RemovePendingExecution removes a pending execution
func (r *Registry) RemovePendingExecution(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.pendingExecutions, id)
}

// ExecutePending executes a pending tool call after confirmation
func (r *Registry) ExecutePending(ctx context.Context, id string) (*ToolResult, error) {
	exec, ok := r.GetPendingExecution(id)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("pending execution %s not found", id),
		}, nil
	}

	result, err := r.Execute(ctx, exec.ToolName, exec.Parameters)
	r.RemovePendingExecution(id)
	return result, err
}
