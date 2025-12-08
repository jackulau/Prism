package routes

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/api/websocket"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/mcp"
	"github.com/jacklau/prism/internal/tools"
	"github.com/jacklau/prism/internal/tools/builtin"
)

// activeGenerations tracks active chat generations for cancellation
var activeGenerations = sync.Map{} // map[conversationID]context.CancelFunc

// handleChatMessage handles incoming chat messages and streams LLM responses
func handleChatMessage(deps *Dependencies, client *websocket.Client, msg *websocket.IncomingMessage) {
	// Validate conversation ID
	if msg.ConversationID == "" {
		client.SendMessage(websocket.NewError("invalid_request", "conversation_id is required"))
		return
	}

	// Get conversation from database
	conversation, err := deps.ConversationRepo.GetByID(msg.ConversationID)
	if err != nil {
		client.SendMessage(websocket.NewError("database_error", "failed to get conversation: "+err.Error()))
		return
	}
	if conversation == nil {
		client.SendMessage(websocket.NewError("not_found", "conversation not found"))
		return
	}

	// Verify the conversation belongs to the user
	if conversation.UserID != client.UserID {
		client.SendMessage(websocket.NewError("forbidden", "not authorized to access this conversation"))
		return
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	activeGenerations.Store(msg.ConversationID, cancel)
	defer func() {
		activeGenerations.Delete(msg.ConversationID)
		cancel()
	}()

	// Save user message to database
	userMsg, err := deps.MessageRepo.Create(msg.ConversationID, "user", msg.Content, nil, "")
	if err != nil {
		log.Printf("Failed to save user message: %v", err)
	}

	// Get message history
	messages, err := deps.MessageRepo.ListByConversationID(msg.ConversationID)
	if err != nil {
		log.Printf("Failed to get message history: %v", err)
		messages = []*repository.Message{}
	}

	// Build LLM messages
	llmMessages := buildLLMMessages(conversation.SystemPrompt, messages, userMsg)

	// Get tools from registry if available
	var toolDefs []llm.ToolDefinition
	if deps.ToolRegistry != nil {
		toolDefs = deps.ToolRegistry.ToLLMTools()
	}

	// Get MCP tools for the user and merge them
	var mcpTools []*mcp.MCPToolWrapper
	if deps.MCPClient != nil {
		mcpTools = mcp.GetMCPToolsForUser(deps.MCPClient, client.UserID)
		if len(mcpTools) > 0 {
			mcpToolDefs := mcp.ToLLMToolDefinitions(mcpTools)
			toolDefs = append(toolDefs, mcpToolDefs...)
			log.Printf("Added %d MCP tools for user %s", len(mcpTools), client.UserID)
		}
	}

	// Create chat request
	req := &llm.ChatRequest{
		Model:    conversation.Model,
		Messages: llmMessages,
		Tools:    toolDefs,
		Stream:   true,
	}

	// Stream response from LLM
	messageID := uuid.New().String()
	streamLLMResponseWithMCP(ctx, deps, client, msg.ConversationID, conversation.Provider, messageID, req, mcpTools)
}

// handleChatStop stops an ongoing chat generation
func handleChatStop(deps *Dependencies, client *websocket.Client, msg *websocket.IncomingMessage) {
	if msg.ConversationID == "" {
		client.SendMessage(websocket.NewError("invalid_request", "conversation_id is required"))
		return
	}

	if cancel, ok := activeGenerations.Load(msg.ConversationID); ok {
		cancel.(context.CancelFunc)()
		activeGenerations.Delete(msg.ConversationID)
		log.Printf("Generation stopped for conversation: %s", msg.ConversationID)
	}

	client.SendMessage(websocket.NewChatComplete(msg.ConversationID, "", "stop"))
}

// handleToolConfirm handles tool confirmation (approve/reject)
func handleToolConfirm(deps *Dependencies, client *websocket.Client, msg *websocket.IncomingMessage) {
	if msg.ExecutionID == "" {
		client.SendMessage(websocket.NewError("invalid_request", "execution_id is required"))
		return
	}

	if deps.ToolRegistry == nil {
		client.SendMessage(websocket.NewError("tool_unavailable", "tool registry not available"))
		return
	}

	// Get pending execution
	pending, ok := deps.ToolRegistry.GetPendingExecution(msg.ExecutionID)
	if !ok {
		client.SendMessage(websocket.NewError("not_found", "pending execution not found"))
		return
	}

	if !msg.Approved {
		// User rejected the tool execution
		deps.ToolRegistry.RemovePendingExecution(msg.ExecutionID)
		client.SendMessage(websocket.NewToolCompleted(pending.ConversationID, msg.ExecutionID, map[string]interface{}{
			"status": "rejected",
			"reason": "User rejected the tool execution",
		}, "rejected"))
		return
	}

	ctx := context.WithValue(context.Background(), builtin.UserIDKey, client.UserID)
	var result interface{}
	var status string

	// Check if this is an MCP tool
	if pending.IsMCPTool {
		// Execute MCP tool
		mcpResult, err := executeMCPTool(ctx, deps, pending)
		deps.ToolRegistry.RemovePendingExecution(msg.ExecutionID)

		if err != nil {
			client.SendMessage(websocket.NewError("mcp_tool_error", err.Error()))
			return
		}

		result = mcpResult
		if mcpResult.Success {
			status = "completed"
		} else {
			status = "failed"
		}
	} else {
		// Execute local tool
		toolResult, err := deps.ToolRegistry.ExecutePending(ctx, msg.ExecutionID)
		if err != nil {
			client.SendMessage(websocket.NewError("tool_error", err.Error()))
			return
		}

		result = toolResult
		if toolResult.Success {
			status = "completed"
		} else {
			status = "failed"
		}
	}

	client.SendMessage(websocket.NewToolCompleted(pending.ConversationID, msg.ExecutionID, result, status))

	// Continue the conversation with the tool result
	continueConversationWithToolResult(ctx, deps, client, pending, result, status)
}

// continueConversationWithToolResult sends the tool result back to the LLM and streams the response
func continueConversationWithToolResult(ctx context.Context, deps *Dependencies, client *websocket.Client, pending *tools.PendingExecution, result interface{}, status string) {
	// Get conversation from database
	conversation, err := deps.ConversationRepo.GetByID(pending.ConversationID)
	if err != nil {
		log.Printf("Failed to get conversation for tool continuation: %v", err)
		return
	}
	if conversation == nil {
		log.Printf("Conversation not found for tool continuation: %s", pending.ConversationID)
		return
	}

	// Serialize the tool result to JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Printf("Failed to marshal tool result: %v", err)
		resultJSON = []byte("{\"error\": \"failed to serialize result\"}")
	}

	// Save the tool result message to database
	// The tool_call_id should reference the original tool call
	_, err = deps.MessageRepo.Create(pending.ConversationID, "tool", string(resultJSON), nil, pending.ID)
	if err != nil {
		log.Printf("Failed to save tool result message: %v", err)
	}

	// Get updated message history
	messages, err := deps.MessageRepo.ListByConversationID(pending.ConversationID)
	if err != nil {
		log.Printf("Failed to get message history for tool continuation: %v", err)
		return
	}

	// Build LLM messages
	llmMessages := buildLLMMessages(conversation.SystemPrompt, messages, nil)

	// Get tools from registry if available
	var toolDefs []llm.ToolDefinition
	if deps.ToolRegistry != nil {
		toolDefs = deps.ToolRegistry.ToLLMTools()
	}

	// Get MCP tools for the user
	var mcpTools []*mcp.MCPToolWrapper
	if deps.MCPClient != nil {
		mcpTools = mcp.GetMCPToolsForUser(deps.MCPClient, client.UserID)
		if len(mcpTools) > 0 {
			mcpToolDefs := mcp.ToLLMToolDefinitions(mcpTools)
			toolDefs = append(toolDefs, mcpToolDefs...)
		}
	}

	// Create chat request
	req := &llm.ChatRequest{
		Model:    conversation.Model,
		Messages: llmMessages,
		Tools:    toolDefs,
		Stream:   true,
	}

	// Stream response from LLM
	messageID := uuid.New().String()
	streamLLMResponseWithMCP(ctx, deps, client, pending.ConversationID, conversation.Provider, messageID, req, mcpTools)
}

// buildLLMMessages converts database messages to LLM messages
func buildLLMMessages(systemPrompt string, history []*repository.Message, userMsg *repository.Message) []llm.Message {
	messages := make([]llm.Message, 0, len(history)+2)

	// Add system prompt if present
	if systemPrompt != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// Add message history
	for _, msg := range history {
		llmMsg := llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}

		// Convert tool calls if present
		if len(msg.ToolCalls) > 0 {
			llmMsg.ToolCalls = make([]llm.ToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				llmMsg.ToolCalls[i] = llm.ToolCall{
					ID:         tc.ID,
					Name:       tc.Name,
					Parameters: tc.Parameters,
				}
			}
		}

		if msg.ToolCallID != "" {
			llmMsg.ToolCallID = msg.ToolCallID
		}

		messages = append(messages, llmMsg)
	}

	return messages
}

// streamLLMResponseWithMCP streams the LLM response to the client with MCP tool support
func streamLLMResponseWithMCP(ctx context.Context, deps *Dependencies, client *websocket.Client, conversationID, provider, messageID string, req *llm.ChatRequest, mcpTools []*mcp.MCPToolWrapper) {
	// Check if provider is set
	if provider == "" {
		client.SendMessage(websocket.NewError("provider_error", "no LLM provider configured for this conversation"))
		return
	}

	// Get the stream from LLM manager
	stream, err := deps.LLMManager.Chat(ctx, provider, req)
	if err != nil {
		client.SendMessage(websocket.NewError("llm_error", "failed to start chat: "+err.Error()))
		return
	}

	var fullResponse strings.Builder
	var finishReason string

	// Build MCP tool lookup map for faster access
	mcpToolMap := make(map[string]*mcp.MCPToolWrapper)
	for _, t := range mcpTools {
		mcpToolMap[t.Name()] = t
	}

	for chunk := range stream {
		// Check if context was cancelled
		select {
		case <-ctx.Done():
			finishReason = "stop"
			goto saveAndComplete
		default:
		}

		if chunk.Error != nil {
			client.SendMessage(websocket.NewError("stream_error", chunk.Error.Error()))
			finishReason = "error"
			goto saveAndComplete
		}

		// Handle text delta
		if chunk.Delta != "" {
			fullResponse.WriteString(chunk.Delta)
			client.SendMessage(websocket.NewChatChunk(conversationID, messageID, chunk.Delta))
		}

		// Handle tool calls
		if len(chunk.ToolCalls) > 0 {
			for _, tc := range chunk.ToolCalls {
				handleToolCallWithMCP(ctx, deps, client, conversationID, messageID, tc, mcpToolMap)
			}
		}

		// Handle finish reason
		if chunk.FinishReason != "" {
			finishReason = chunk.FinishReason
		}
	}

saveAndComplete:
	// Save assistant message to database
	if fullResponse.Len() > 0 {
		_, err := deps.MessageRepo.Create(conversationID, "assistant", fullResponse.String(), nil, "")
		if err != nil {
			log.Printf("Failed to save assistant message: %v", err)
		}
	}

	// Send completion message
	if finishReason == "" {
		finishReason = "stop"
	}
	client.SendMessage(websocket.NewChatComplete(conversationID, messageID, finishReason))

	// Track completion
	if deps.IntegrationManager != nil {
		deps.IntegrationManager.TrackChatCompleted(client.UserID, conversationID, messageID, finishReason)
	}
}

// handleToolCall handles a tool call from the LLM
func handleToolCall(ctx context.Context, deps *Dependencies, client *websocket.Client, conversationID, messageID string, tc llm.ToolCall) {
	if deps.ToolRegistry == nil {
		client.SendMessage(websocket.NewError("tool_unavailable", "tool registry not available"))
		return
	}

	tool, ok := deps.ToolRegistry.Get(tc.Name)
	if !ok {
		client.SendMessage(websocket.NewError("tool_not_found", "tool not found: "+tc.Name))
		return
	}

	executionID := uuid.New().String()

	// Check if tool requires confirmation
	if tool.RequiresConfirmation() {
		// Store pending execution
		pending := &tools.PendingExecution{
			ID:             executionID,
			ToolName:       tc.Name,
			Parameters:     tc.Parameters,
			ConversationID: conversationID,
			MessageID:      messageID,
			UserID:         client.UserID,
		}
		deps.ToolRegistry.AddPendingExecution(pending)

		// Send confirmation request to client
		client.SendMessage(&websocket.OutgoingMessage{
			Type:           websocket.TypeToolConfirm,
			ConversationID: conversationID,
			ExecutionID:    executionID,
			ToolName:       tc.Name,
			Parameters:     tc.Parameters,
		})
		return
	}

	// Execute tool immediately (no confirmation needed)
	client.SendMessage(websocket.NewToolStarted(conversationID, executionID, tc.Name, tc.Parameters))

	// Add user ID to context
	toolCtx := context.WithValue(ctx, builtin.UserIDKey, client.UserID)
	result, err := deps.ToolRegistry.Execute(toolCtx, tc.Name, tc.Parameters)
	if err != nil {
		client.SendMessage(websocket.NewError("tool_error", err.Error()))
		return
	}

	status := "completed"
	if !result.Success {
		status = "failed"
	}

	client.SendMessage(websocket.NewToolCompleted(conversationID, executionID, result, status))
}

// handleToolCallWithMCP handles a tool call, routing to either local tools or MCP tools
func handleToolCallWithMCP(ctx context.Context, deps *Dependencies, client *websocket.Client, conversationID, messageID string, tc llm.ToolCall, mcpToolMap map[string]*mcp.MCPToolWrapper) {
	executionID := uuid.New().String()

	// Check if this is an MCP tool
	if mcpTool, isMCP := mcpToolMap[tc.Name]; isMCP {
		handleMCPToolCall(ctx, deps, client, conversationID, messageID, executionID, mcpTool, tc.Parameters)
		return
	}

	// Fall back to local tool registry
	if deps.ToolRegistry == nil {
		client.SendMessage(websocket.NewError("tool_unavailable", "tool registry not available"))
		return
	}

	tool, ok := deps.ToolRegistry.Get(tc.Name)
	if !ok {
		client.SendMessage(websocket.NewError("tool_not_found", "tool not found: "+tc.Name))
		return
	}

	// Check if tool requires confirmation
	if tool.RequiresConfirmation() {
		// Store pending execution
		pending := &tools.PendingExecution{
			ID:             executionID,
			ToolName:       tc.Name,
			Parameters:     tc.Parameters,
			ConversationID: conversationID,
			MessageID:      messageID,
			UserID:         client.UserID,
		}
		deps.ToolRegistry.AddPendingExecution(pending)

		// Send confirmation request to client
		client.SendMessage(&websocket.OutgoingMessage{
			Type:           websocket.TypeToolConfirm,
			ConversationID: conversationID,
			ExecutionID:    executionID,
			ToolName:       tc.Name,
			Parameters:     tc.Parameters,
		})
		return
	}

	// Execute tool immediately (no confirmation needed)
	client.SendMessage(websocket.NewToolStarted(conversationID, executionID, tc.Name, tc.Parameters))

	// Add user ID to context
	toolCtx := context.WithValue(ctx, builtin.UserIDKey, client.UserID)
	result, err := deps.ToolRegistry.Execute(toolCtx, tc.Name, tc.Parameters)
	if err != nil {
		client.SendMessage(websocket.NewError("tool_error", err.Error()))
		return
	}

	status := "completed"
	if !result.Success {
		status = "failed"
	}

	client.SendMessage(websocket.NewToolCompleted(conversationID, executionID, result, status))
}

// handleMCPToolCall handles execution of an MCP tool
func handleMCPToolCall(ctx context.Context, deps *Dependencies, client *websocket.Client, conversationID, messageID, executionID string, mcpTool *mcp.MCPToolWrapper, params map[string]interface{}) {
	// MCP tools always require confirmation for security
	pending := &tools.PendingExecution{
		ID:             executionID,
		ToolName:       mcpTool.Name(),
		Parameters:     params,
		ConversationID: conversationID,
		MessageID:      messageID,
		UserID:         client.UserID,
		IsMCPTool:      true,
		MCPServerID:    mcpTool.ServerID(),
		MCPToolName:    mcpTool.OriginalName(),
	}

	if deps.ToolRegistry != nil {
		deps.ToolRegistry.AddPendingExecution(pending)
	}

	// Send confirmation request to client with MCP indicator
	client.SendMessage(&websocket.OutgoingMessage{
		Type:           websocket.TypeToolConfirm,
		ConversationID: conversationID,
		ExecutionID:    executionID,
		ToolName:       mcpTool.Name(),
		Parameters:     params,
		IsMCPTool:      true,
		MCPServerName:  mcpTool.Description(), // Contains server name in description
	})
}

// executeMCPTool executes an MCP tool after user confirmation
func executeMCPTool(ctx context.Context, deps *Dependencies, pending *tools.PendingExecution) (*tools.ExecutionResult, error) {
	if deps.MCPClient == nil {
		return &tools.ExecutionResult{
			Success: false,
			Error:   "MCP client not available",
		}, nil
	}

	result, err := deps.MCPClient.ExecuteTool(ctx, pending.MCPServerID, pending.MCPToolName, pending.Parameters)
	if err != nil {
		return &tools.ExecutionResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &tools.ExecutionResult{
		Success: true,
		Data:    result,
	}, nil
}
