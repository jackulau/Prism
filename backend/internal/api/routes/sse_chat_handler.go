package routes

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jacklau/prism/internal/api/sse"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/mcp"
	"github.com/jacklau/prism/internal/tools"
	"github.com/jacklau/prism/internal/tools/builtin"
)

// SSEChatRequest represents a chat request via SSE
type SSEChatRequest struct {
	ConversationID   string                 `json:"conversation_id"`
	Content          string                 `json:"content"`
	Attachments      []Attachment           `json:"attachments,omitempty"`
	Mode             string                 `json:"mode,omitempty"`
	ExtendedThinking bool                   `json:"extended_thinking,omitempty"`
	FileContext      *FileContext           `json:"file_context,omitempty"`
	Params           map[string]interface{} `json:"params,omitempty"`
}

// Attachment represents a file attachment
type Attachment struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
}

// FileContext represents file context for chat messages
type FileContext struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Language string `json:"language,omitempty"`
}

// HandleSSEChat handles chat messages via SSE
// POST /api/v1/sse/chat
func HandleSSEChat(deps *Dependencies, sseService *sse.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user from auth middleware
		userID, ok := c.Locals("userID").(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		// Parse request
		var req SSEChatRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body: " + err.Error(),
			})
		}

		// Validate conversation ID
		if req.ConversationID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "conversation_id is required",
			})
		}

		// Get conversation from database
		conversation, err := deps.ConversationRepo.GetByID(req.ConversationID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get conversation: " + err.Error(),
			})
		}
		if conversation == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "conversation not found",
			})
		}

		// Verify the conversation belongs to the user
		if conversation.UserID != userID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "not authorized to access this conversation",
			})
		}

		// Create cancellable context
		ctx, cancel := context.WithCancel(context.Background())
		activeGenerations.Store(req.ConversationID, cancel)
		defer func() {
			activeGenerations.Delete(req.ConversationID)
			cancel()
		}()

		// Reset iteration count for new user message
		resetIterationCount(req.ConversationID)

		// Save user message to database
		userMsg, err := deps.MessageRepo.Create(req.ConversationID, "user", req.Content, nil, "")
		if err != nil {
			log.Printf("Failed to save user message: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to save message",
			})
		}

		// Get message history
		messages, err := deps.MessageRepo.ListByConversationID(req.ConversationID)
		if err != nil {
			log.Printf("Failed to get message history: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get message history",
			})
		}

		// Build LLM messages
		llmMessages := buildLLMMessages(conversation.SystemPrompt, messages, userMsg)

		// Get tools from registry if available
		var toolDefs []llm.ToolDefinition
		if deps.ToolRegistry != nil {
			toolDefs = deps.ToolRegistry.ToLLMTools()
		}

		// Get HTTP MCP tools for the user and merge them
		var mcpTools []*mcp.MCPToolWrapper
		if deps.MCPClient != nil {
			mcpTools = mcp.GetMCPToolsForUser(deps.MCPClient, userID)
			if len(mcpTools) > 0 {
				mcpToolDefs := mcp.ToLLMToolDefinitions(mcpTools)
				toolDefs = append(toolDefs, mcpToolDefs...)
				log.Printf("Added %d HTTP MCP tools for user %s", len(mcpTools), userID)
			}
		}

		// Get stdio MCP tools for the user and merge them
		var stdioMCPTools []*mcp.StdioMCPToolWrapper
		if deps.StdioMCPClient != nil {
			stdioMCPTools = mcp.GetStdioMCPToolsForUser(deps.StdioMCPClient, userID)
			if len(stdioMCPTools) > 0 {
				stdioToolDefs := mcp.StdioToLLMToolDefinitions(stdioMCPTools)
				toolDefs = append(toolDefs, stdioToolDefs...)
				log.Printf("Added %d stdio MCP tools for user %s", len(stdioMCPTools), userID)
			}
		}

		// Create chat request
		llmReq := &llm.ChatRequest{
			Model:    conversation.Model,
			Messages: llmMessages,
			Tools:    toolDefs,
			Stream:   true,
		}

		// Generate message ID
		messageID := uuid.New().String()

		// Stream response from LLM via SSE
		streamLLMResponseViaSSE(ctx, deps, sseService, userID, req.ConversationID, conversation.Provider, messageID, llmReq, mcpTools, stdioMCPTools)

		// Return success - actual response is streamed via SSE
		return c.JSON(fiber.Map{
			"message_id": messageID,
			"status":     "streaming",
		})
	}
}

// streamLLMResponseViaSSE streams the LLM response to the client via SSE
func streamLLMResponseViaSSE(ctx context.Context, deps *Dependencies, sseService *sse.Service, userID, conversationID, provider, messageID string, req *llm.ChatRequest, mcpTools []*mcp.MCPToolWrapper, stdioMCPTools []*mcp.StdioMCPToolWrapper) {
	// Check if provider is set
	if provider == "" {
		sseService.BroadcastToUser(userID, sse.NewError("provider_error", "no LLM provider configured for this conversation"))
		return
	}

	// Load user's API key from database for providers that require it
	if provider != "ollama" && deps.ProviderKeyRepo != nil && deps.EncryptionService != nil {
		providerKey, err := deps.ProviderKeyRepo.GetKey(userID, provider)
		if err == nil && providerKey != nil {
			decryptedKey, err := deps.EncryptionService.Decrypt(providerKey.EncryptedKey, providerKey.KeyNonce)
			if err == nil {
				deps.LLMManager.SetAPIKey(provider, string(decryptedKey))
			}
		}
	}

	// Check if provider has a valid API key configured
	if !deps.LLMManager.HasValidKey(provider) {
		sseService.BroadcastToUser(userID, sse.NewError("api_key_missing",
			"API key not configured for provider: "+provider+". Please add your API key in Settings."))
		return
	}

	// Get the stream from LLM manager
	stream, err := deps.LLMManager.Chat(ctx, provider, req)
	if err != nil {
		sseService.BroadcastToUser(userID, sse.NewError("llm_error", "failed to start chat: "+err.Error()))
		return
	}

	var fullResponse strings.Builder
	var finishReason string
	var collectedToolCalls []llm.ToolCall

	// Build HTTP MCP tool lookup map
	mcpToolMap := make(map[string]*mcp.MCPToolWrapper)
	for _, t := range mcpTools {
		mcpToolMap[t.Name()] = t
	}

	// Build stdio MCP tool lookup map
	stdioMCPToolMap := make(map[string]*mcp.StdioMCPToolWrapper)
	for _, t := range stdioMCPTools {
		stdioMCPToolMap[t.Name()] = t
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
			sseService.BroadcastToUser(userID, sse.NewError("stream_error", chunk.Error.Error()))
			finishReason = "error"
			goto saveAndComplete
		}

		// Handle text delta
		if chunk.Delta != "" {
			fullResponse.WriteString(chunk.Delta)
			sseService.BroadcastToUser(userID, sse.NewChatChunk(conversationID, messageID, chunk.Delta))
		}

		// Handle tool calls
		if len(chunk.ToolCalls) > 0 {
			for _, tc := range chunk.ToolCalls {
				collectedToolCalls = append(collectedToolCalls, tc)
				handleToolCallViaSSE(ctx, deps, sseService, userID, conversationID, messageID, tc, mcpToolMap, stdioMCPToolMap)
			}
		}

		// Handle finish reason
		if chunk.FinishReason != "" {
			finishReason = chunk.FinishReason
		}
	}

saveAndComplete:
	// Save assistant message to database
	if fullResponse.Len() > 0 || len(collectedToolCalls) > 0 {
		toolCalls := convertToRepoToolCalls(collectedToolCalls)
		_, err := deps.MessageRepo.Create(conversationID, "assistant", fullResponse.String(), toolCalls, "")
		if err != nil {
			log.Printf("Failed to save assistant message: %v", err)
		}
	}

	// Send completion message
	if finishReason == "" {
		finishReason = "stop"
	}
	sseService.BroadcastToUser(userID, sse.NewChatComplete(conversationID, messageID, finishReason))

	// Track completion
	if deps.IntegrationManager != nil {
		deps.IntegrationManager.TrackChatCompleted(userID, conversationID, messageID, finishReason)
	}
}

// handleToolCallViaSSE handles a tool call, routing to local tools, HTTP MCP tools, or stdio MCP tools via SSE
func handleToolCallViaSSE(ctx context.Context, deps *Dependencies, sseService *sse.Service, userID, conversationID, messageID string, tc llm.ToolCall, mcpToolMap map[string]*mcp.MCPToolWrapper, stdioMCPToolMap map[string]*mcp.StdioMCPToolWrapper) {
	executionID := uuid.New().String()

	// Check if this is an HTTP MCP tool
	if mcpTool, isMCP := mcpToolMap[tc.Name]; isMCP {
		handleHTTPMCPToolCallViaSSE(ctx, deps, sseService, userID, conversationID, messageID, executionID, tc.ID, mcpTool, tc.Parameters)
		return
	}

	// Check if this is a stdio MCP tool
	if stdioMCPTool, isStdioMCP := stdioMCPToolMap[tc.Name]; isStdioMCP {
		handleStdioMCPToolCallViaSSE(ctx, deps, sseService, userID, conversationID, messageID, executionID, tc.ID, stdioMCPTool, tc.Parameters)
		return
	}

	// Fall back to local tool registry
	if deps.ToolRegistry == nil {
		sseService.BroadcastToUser(userID, sse.NewError("tool_unavailable", "tool registry not available"))
		return
	}

	tool, ok := deps.ToolRegistry.Get(tc.Name)
	if !ok {
		sseService.BroadcastToUser(userID, sse.NewError("tool_not_found", "tool not found: "+tc.Name))
		return
	}

	// Check if tool requires confirmation
	if tool.RequiresConfirmation() {
		pending := &tools.PendingExecution{
			ID:             executionID,
			ToolCallID:     tc.ID,
			ToolName:       tc.Name,
			Parameters:     tc.Parameters,
			ConversationID: conversationID,
			MessageID:      messageID,
			UserID:         userID,
		}
		deps.ToolRegistry.AddPendingExecution(pending)

		// Send confirmation request via SSE
		sseService.BroadcastToUser(userID, sse.NewToolConfirm(conversationID, executionID, tc.Name, tc.Parameters))
		return
	}

	// Execute tool immediately
	sseService.BroadcastToUser(userID, sse.NewToolStarted(conversationID, executionID, tc.Name, tc.Parameters))

	toolCtx := context.WithValue(ctx, builtin.UserIDKey, userID)
	result, err := deps.ToolRegistry.Execute(toolCtx, tc.Name, tc.Parameters)
	if err != nil {
		sseService.BroadcastToUser(userID, sse.NewError("tool_error", err.Error()))
		return
	}

	status := "completed"
	if !result.Success {
		status = "failed"
	}

	sseService.BroadcastToUser(userID, sse.NewToolCompleted(conversationID, executionID, result, status))

	// Continue conversation with tool result
	pending := &tools.PendingExecution{
		ID:             executionID,
		ToolCallID:     tc.ID,
		ToolName:       tc.Name,
		Parameters:     tc.Parameters,
		ConversationID: conversationID,
		MessageID:      messageID,
		UserID:         userID,
	}
	continueConversationWithToolResultViaSSE(ctx, deps, sseService, userID, pending, result, status)
}

// handleHTTPMCPToolCallViaSSE handles execution of an HTTP MCP tool via SSE
func handleHTTPMCPToolCallViaSSE(ctx context.Context, deps *Dependencies, sseService *sse.Service, userID, conversationID, messageID, executionID, toolCallID string, mcpTool *mcp.MCPToolWrapper, params map[string]interface{}) {
	approvalConfig := getDefaultAutoApprovalConfig()
	iterationCount := incrementIterationCount(conversationID)

	if approvalConfig.ShouldCheckIn(iterationCount) {
		sseService.BroadcastToUser(userID, sse.NewEvent(sse.EventAgentCheckIn, map[string]interface{}{
			"conversation_id": conversationID,
			"iteration_count": iterationCount,
			"message":         "Agent has reached the maximum number of tool executions. Would you like to continue?",
		}))
		return
	}

	if approvalConfig.ShouldAutoApprove(mcpTool.Name(), true) {
		// Execute immediately
		sseService.BroadcastToUser(userID, sse.NewEvent(sse.EventToolStarted, sse.ToolStartedData{
			ConversationID: conversationID,
			ExecutionID:    executionID,
			ToolName:       mcpTool.Name(),
			Parameters:     params,
			IsMCPTool:      true,
			MCPServerName:  mcpTool.Description(),
		}))

		result, err := deps.MCPClient.ExecuteTool(ctx, mcpTool.ServerID(), mcpTool.OriginalName(), params)

		var execResult *tools.ExecutionResult
		if err != nil {
			execResult = &tools.ExecutionResult{Success: false, Error: err.Error()}
		} else {
			execResult = &tools.ExecutionResult{Success: true, Data: result}
		}

		status := "completed"
		if !execResult.Success {
			status = "failed"
		}
		sseService.BroadcastToUser(userID, sse.NewToolCompleted(conversationID, executionID, execResult, status))

		pending := &tools.PendingExecution{
			ID:             executionID,
			ToolCallID:     toolCallID,
			ToolName:       mcpTool.Name(),
			Parameters:     params,
			ConversationID: conversationID,
			MessageID:      messageID,
			UserID:         userID,
			IsMCPTool:      true,
			IterationCount: iterationCount,
		}
		continueConversationWithToolResultViaSSE(ctx, deps, sseService, userID, pending, execResult, status)
		return
	}

	// Require confirmation
	pending := &tools.PendingExecution{
		ID:             executionID,
		ToolCallID:     toolCallID,
		ToolName:       mcpTool.Name(),
		Parameters:     params,
		ConversationID: conversationID,
		MessageID:      messageID,
		UserID:         userID,
		IsMCPTool:      true,
		IsStdioMCP:     false,
		MCPServerID:    mcpTool.ServerID(),
		MCPToolName:    mcpTool.OriginalName(),
		IterationCount: iterationCount,
	}

	if deps.ToolRegistry != nil {
		deps.ToolRegistry.AddPendingExecution(pending)
	}

	sseService.BroadcastToUser(userID, sse.NewEvent(sse.EventToolConfirm, sse.ToolConfirmData{
		ConversationID: conversationID,
		ExecutionID:    executionID,
		ToolName:       mcpTool.Name(),
		Parameters:     params,
		IsMCPTool:      true,
		MCPServerName:  mcpTool.Description(),
		IterationCount: iterationCount,
	}))
}

// handleStdioMCPToolCallViaSSE handles execution of a stdio MCP tool via SSE
func handleStdioMCPToolCallViaSSE(ctx context.Context, deps *Dependencies, sseService *sse.Service, userID, conversationID, messageID, executionID, toolCallID string, mcpTool *mcp.StdioMCPToolWrapper, params map[string]interface{}) {
	approvalConfig := getDefaultAutoApprovalConfig()
	iterationCount := incrementIterationCount(conversationID)

	if approvalConfig.ShouldCheckIn(iterationCount) {
		sseService.BroadcastToUser(userID, sse.NewEvent(sse.EventAgentCheckIn, map[string]interface{}{
			"conversation_id": conversationID,
			"iteration_count": iterationCount,
			"message":         "Agent has reached the maximum number of tool executions. Would you like to continue?",
		}))
		return
	}

	if approvalConfig.ShouldAutoApprove(mcpTool.Name(), true) {
		// Execute immediately
		sseService.BroadcastToUser(userID, sse.NewEvent(sse.EventToolStarted, sse.ToolStartedData{
			ConversationID: conversationID,
			ExecutionID:    executionID,
			ToolName:       mcpTool.Name(),
			Parameters:     params,
			IsMCPTool:      true,
			IsStdioMCP:     true,
			MCPServerName:  mcpTool.Description(),
		}))

		result, err := deps.StdioMCPClient.ExecuteTool(ctx, mcpTool.ServerID(), mcpTool.OriginalName(), params)

		var execResult *tools.ExecutionResult
		if err != nil {
			execResult = &tools.ExecutionResult{Success: false, Error: err.Error()}
		} else {
			execResult = &tools.ExecutionResult{Success: true, Data: result}
		}

		status := "completed"
		if !execResult.Success {
			status = "failed"
		}
		sseService.BroadcastToUser(userID, sse.NewToolCompleted(conversationID, executionID, execResult, status))

		pending := &tools.PendingExecution{
			ID:             executionID,
			ToolCallID:     toolCallID,
			ToolName:       mcpTool.Name(),
			Parameters:     params,
			ConversationID: conversationID,
			MessageID:      messageID,
			UserID:         userID,
			IsMCPTool:      true,
			IsStdioMCP:     true,
			IterationCount: iterationCount,
		}
		continueConversationWithToolResultViaSSE(ctx, deps, sseService, userID, pending, execResult, status)
		return
	}

	// Require confirmation
	pending := &tools.PendingExecution{
		ID:             executionID,
		ToolCallID:     toolCallID,
		ToolName:       mcpTool.Name(),
		Parameters:     params,
		ConversationID: conversationID,
		MessageID:      messageID,
		UserID:         userID,
		IsMCPTool:      true,
		IsStdioMCP:     true,
		MCPServerID:    mcpTool.ServerID(),
		MCPToolName:    mcpTool.OriginalName(),
		IterationCount: iterationCount,
	}

	if deps.ToolRegistry != nil {
		deps.ToolRegistry.AddPendingExecution(pending)
	}

	sseService.BroadcastToUser(userID, sse.NewEvent(sse.EventToolConfirm, sse.ToolConfirmData{
		ConversationID: conversationID,
		ExecutionID:    executionID,
		ToolName:       mcpTool.Name(),
		Parameters:     params,
		IsMCPTool:      true,
		IsStdioMCP:     true,
		MCPServerName:  mcpTool.Description(),
		IterationCount: iterationCount,
	}))
}

// continueConversationWithToolResultViaSSE sends the tool result back to the LLM and streams the response via SSE
func continueConversationWithToolResultViaSSE(ctx context.Context, deps *Dependencies, sseService *sse.Service, userID string, pending *tools.PendingExecution, result interface{}, status string) {
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
	_, err = deps.MessageRepo.Create(pending.ConversationID, "tool", string(resultJSON), nil, pending.ToolCallID)
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

	// Get HTTP MCP tools for the user
	var mcpTools []*mcp.MCPToolWrapper
	if deps.MCPClient != nil {
		mcpTools = mcp.GetMCPToolsForUser(deps.MCPClient, userID)
		if len(mcpTools) > 0 {
			mcpToolDefs := mcp.ToLLMToolDefinitions(mcpTools)
			toolDefs = append(toolDefs, mcpToolDefs...)
		}
	}

	// Get stdio MCP tools for the user
	var stdioMCPTools []*mcp.StdioMCPToolWrapper
	if deps.StdioMCPClient != nil {
		stdioMCPTools = mcp.GetStdioMCPToolsForUser(deps.StdioMCPClient, userID)
		if len(stdioMCPTools) > 0 {
			stdioToolDefs := mcp.StdioToLLMToolDefinitions(stdioMCPTools)
			toolDefs = append(toolDefs, stdioToolDefs...)
		}
	}

	// Create chat request
	req := &llm.ChatRequest{
		Model:    conversation.Model,
		Messages: llmMessages,
		Tools:    toolDefs,
		Stream:   true,
	}

	// Stream response from LLM via SSE
	messageID := uuid.New().String()
	streamLLMResponseViaSSE(ctx, deps, sseService, userID, pending.ConversationID, conversation.Provider, messageID, req, mcpTools, stdioMCPTools)
}

// HandleSSEToolConfirm handles tool confirmation via SSE
// POST /api/v1/sse/tool/confirm
func HandleSSEToolConfirm(deps *Dependencies, sseService *sse.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		type ConfirmRequest struct {
			ExecutionID string `json:"execution_id"`
			Approved    bool   `json:"approved"`
		}

		var req ConfirmRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		if req.ExecutionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "execution_id is required",
			})
		}

		if deps.ToolRegistry == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "tool registry not available",
			})
		}

		pending, ok := deps.ToolRegistry.GetPendingExecution(req.ExecutionID)
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "pending execution not found",
			})
		}

		if !req.Approved {
			deps.ToolRegistry.RemovePendingExecution(req.ExecutionID)
			sseService.BroadcastToUser(userID, sse.NewToolCompleted(pending.ConversationID, req.ExecutionID, map[string]interface{}{
				"status": "rejected",
				"reason": "User rejected the tool execution",
			}, "rejected"))
			return c.JSON(fiber.Map{"status": "rejected"})
		}

		ctx := context.WithValue(context.Background(), builtin.UserIDKey, userID)
		var result interface{}
		var status string

		if pending.IsMCPTool {
			mcpResult, err := executeMCPTool(ctx, deps, pending)
			deps.ToolRegistry.RemovePendingExecution(req.ExecutionID)

			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": err.Error(),
				})
			}

			result = mcpResult
			if mcpResult.Success {
				status = "completed"
			} else {
				status = "failed"
			}
		} else {
			toolResult, err := deps.ToolRegistry.ExecutePending(ctx, req.ExecutionID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": err.Error(),
				})
			}

			result = toolResult
			if toolResult.Success {
				status = "completed"
			} else {
				status = "failed"
			}
		}

		sseService.BroadcastToUser(userID, sse.NewToolCompleted(pending.ConversationID, req.ExecutionID, result, status))

		// Continue the conversation with the tool result
		continueConversationWithToolResultViaSSE(ctx, deps, sseService, userID, pending, result, status)

		return c.JSON(fiber.Map{"status": status})
	}
}

// HandleSSEChatStop handles stopping an ongoing chat generation via SSE
// POST /api/v1/sse/chat/stop
func HandleSSEChatStop(deps *Dependencies, sseService *sse.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		type StopRequest struct {
			ConversationID string `json:"conversation_id"`
		}

		var req StopRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		if req.ConversationID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "conversation_id is required",
			})
		}

		if cancel, ok := activeGenerations.Load(req.ConversationID); ok {
			cancel.(context.CancelFunc)()
			activeGenerations.Delete(req.ConversationID)
			log.Printf("SSE Generation stopped for conversation: %s", req.ConversationID)
		}

		sseService.BroadcastToUser(userID, sse.NewChatComplete(req.ConversationID, "", "stop"))

		return c.JSON(fiber.Map{"status": "stopped"})
	}
}
