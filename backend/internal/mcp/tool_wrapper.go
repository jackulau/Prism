package mcp

import (
	"context"
	"fmt"

	"github.com/jacklau/prism/internal/llm"
)

// MCPToolWrapper wraps a remote MCP tool to be used in the local tool registry
type MCPToolWrapper struct {
	client      *Client
	serverID    string
	serverName  string
	toolName    string
	description string
	parameters  llm.JSONSchema
}

// NewMCPToolWrapper creates a new MCP tool wrapper
func NewMCPToolWrapper(client *Client, serverID, serverName string, tool *RemoteTool) *MCPToolWrapper {
	return &MCPToolWrapper{
		client:      client,
		serverID:    serverID,
		serverName:  serverName,
		toolName:    tool.Name,
		description: tool.Description,
		parameters:  tool.Parameters,
	}
}

// Name returns the tool name with server prefix to avoid conflicts
func (t *MCPToolWrapper) Name() string {
	return fmt.Sprintf("mcp_%s_%s", t.serverName, t.toolName)
}

// OriginalName returns the original tool name without prefix
func (t *MCPToolWrapper) OriginalName() string {
	return t.toolName
}

// ServerID returns the MCP server ID
func (t *MCPToolWrapper) ServerID() string {
	return t.serverID
}

// Description returns the tool description
func (t *MCPToolWrapper) Description() string {
	return fmt.Sprintf("[MCP: %s] %s", t.serverName, t.description)
}

// Parameters returns the tool parameters schema
func (t *MCPToolWrapper) Parameters() llm.JSONSchema {
	return t.parameters
}

// Execute executes the tool on the remote MCP server
func (t *MCPToolWrapper) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	result, err := t.client.ExecuteTool(ctx, t.serverID, t.toolName, params)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"result":  result,
	}, nil
}

// RequiresConfirmation returns whether the tool requires user confirmation
// MCP tools always require confirmation for security
func (t *MCPToolWrapper) RequiresConfirmation() bool {
	return true
}

// IsMCPTool returns true to identify this as an MCP tool
func (t *MCPToolWrapper) IsMCPTool() bool {
	return true
}

// GetMCPToolsForUser returns all MCP tools available to a user
func GetMCPToolsForUser(client *Client, userID string) []*MCPToolWrapper {
	if client == nil {
		return nil
	}

	remoteTools := client.GetAllTools(userID)
	wrappers := make([]*MCPToolWrapper, 0, len(remoteTools))

	for _, rt := range remoteTools {
		wrapper := &MCPToolWrapper{
			client:      client,
			serverID:    rt.ServerID,
			serverName:  rt.ServerName,
			toolName:    rt.Name,
			description: rt.Description,
			parameters:  rt.Parameters,
		}
		wrappers = append(wrappers, wrapper)
	}

	return wrappers
}

// ToLLMToolDefinitions converts MCP tools to LLM tool definitions
func ToLLMToolDefinitions(tools []*MCPToolWrapper) []llm.ToolDefinition {
	defs := make([]llm.ToolDefinition, len(tools))
	for i, t := range tools {
		defs[i] = llm.ToolDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Parameters(),
		}
	}
	return defs
}

// StdioMCPToolWrapper wraps a stdio MCP tool to be used in the local tool registry
type StdioMCPToolWrapper struct {
	client      *StdioClient
	serverID    string
	serverName  string
	toolName    string
	description string
	parameters  llm.JSONSchema
}

// Name returns the tool name with server prefix to avoid conflicts
func (t *StdioMCPToolWrapper) Name() string {
	return fmt.Sprintf("mcp_%s_%s", t.serverName, t.toolName)
}

// OriginalName returns the original tool name without prefix
func (t *StdioMCPToolWrapper) OriginalName() string {
	return t.toolName
}

// ServerID returns the MCP server ID
func (t *StdioMCPToolWrapper) ServerID() string {
	return t.serverID
}

// Description returns the tool description
func (t *StdioMCPToolWrapper) Description() string {
	return fmt.Sprintf("[MCP: %s] %s", t.serverName, t.description)
}

// Parameters returns the tool parameters schema
func (t *StdioMCPToolWrapper) Parameters() llm.JSONSchema {
	return t.parameters
}

// Execute executes the tool on the stdio MCP server
func (t *StdioMCPToolWrapper) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	result, err := t.client.ExecuteTool(ctx, t.serverID, t.toolName, params)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"result":  result,
	}, nil
}

// RequiresConfirmation returns whether the tool requires user confirmation
// MCP tools always require confirmation for security
func (t *StdioMCPToolWrapper) RequiresConfirmation() bool {
	return true
}

// IsMCPTool returns true to identify this as an MCP tool
func (t *StdioMCPToolWrapper) IsMCPTool() bool {
	return true
}

// IsStdioMCP returns true to identify this as a stdio MCP tool
func (t *StdioMCPToolWrapper) IsStdioMCP() bool {
	return true
}

// GetStdioMCPToolsForUser returns all stdio MCP tools available to a user
func GetStdioMCPToolsForUser(client *StdioClient, userID string) []*StdioMCPToolWrapper {
	if client == nil {
		return nil
	}

	remoteTools := client.GetAllTools(userID)
	wrappers := make([]*StdioMCPToolWrapper, 0, len(remoteTools))

	for _, rt := range remoteTools {
		wrapper := &StdioMCPToolWrapper{
			client:      client,
			serverID:    rt.ServerID,
			serverName:  rt.ServerName,
			toolName:    rt.Name,
			description: rt.Description,
			parameters:  rt.Parameters,
		}
		wrappers = append(wrappers, wrapper)
	}

	return wrappers
}

// StdioToLLMToolDefinitions converts stdio MCP tools to LLM tool definitions
func StdioToLLMToolDefinitions(tools []*StdioMCPToolWrapper) []llm.ToolDefinition {
	defs := make([]llm.ToolDefinition, len(tools))
	for i, t := range tools {
		defs[i] = llm.ToolDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Parameters(),
		}
	}
	return defs
}
