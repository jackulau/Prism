package agent

import (
	"context"

	"github.com/jacklau/prism/internal/types"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	// ContextKeyAgentID is the context key for agent ID
	ContextKeyAgentID contextKey = "agent_id"
	// ContextKeyAgentName is the context key for agent name
	ContextKeyAgentName contextKey = "agent_name"
	// ContextKeyAgentType is the context key for agent type
	ContextKeyAgentType contextKey = "agent_type"
	// ContextKeyMessageID is the context key for message ID
	ContextKeyMessageID contextKey = "message_id"
	// ContextKeyConversationID is the context key for conversation ID
	ContextKeyConversationID contextKey = "conversation_id"
	// ContextKeyWorkflowID is the context key for workflow ID
	ContextKeyWorkflowID contextKey = "workflow_id"
	// ContextKeyStepID is the context key for workflow step ID
	ContextKeyStepID contextKey = "step_id"
	// ContextKeyUserID is the context key for user ID
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyToolName is the context key for the currently executing tool name
	ContextKeyToolName contextKey = "tool_name"
	// ContextKeyToolSlug is the context key for the currently executing tool slug
	ContextKeyToolSlug contextKey = "tool_slug"
)

// Agent type constants
const (
	AgentTypeAssistant  = "assistant"
	AgentTypeAutonomous = "autonomous"
	AgentTypeWorkflow   = "workflow"
)

// WithAgentContext adds agent information to the context
func WithAgentContext(ctx context.Context, agentID, agentName, agentType string) context.Context {
	ctx = context.WithValue(ctx, ContextKeyAgentID, agentID)
	ctx = context.WithValue(ctx, ContextKeyAgentName, agentName)
	ctx = context.WithValue(ctx, ContextKeyAgentType, agentType)
	return ctx
}

// GetAgentFromContext extracts agent information from the context
func GetAgentFromContext(ctx context.Context) (agentID, agentName, agentType string, ok bool) {
	agentID, idOk := ctx.Value(ContextKeyAgentID).(string)
	agentName, nameOk := ctx.Value(ContextKeyAgentName).(string)
	agentType, typeOk := ctx.Value(ContextKeyAgentType).(string)

	ok = idOk && agentID != ""
	if nameOk {
		ok = ok || agentName != ""
	}
	if typeOk {
		ok = ok || agentType != ""
	}

	return agentID, agentName, agentType, ok
}

// WithMessageContext adds message/conversation information to the context
func WithMessageContext(ctx context.Context, messageID, conversationID string) context.Context {
	ctx = context.WithValue(ctx, ContextKeyMessageID, messageID)
	ctx = context.WithValue(ctx, ContextKeyConversationID, conversationID)
	return ctx
}

// GetMessageFromContext extracts message information from the context
func GetMessageFromContext(ctx context.Context) (messageID, conversationID string, ok bool) {
	messageID, msgOk := ctx.Value(ContextKeyMessageID).(string)
	conversationID, convOk := ctx.Value(ContextKeyConversationID).(string)
	ok = (msgOk && messageID != "") || (convOk && conversationID != "")
	return messageID, conversationID, ok
}

// WithWorkflowContext adds workflow information to the context
func WithWorkflowContext(ctx context.Context, workflowID, stepID string) context.Context {
	ctx = context.WithValue(ctx, ContextKeyWorkflowID, workflowID)
	ctx = context.WithValue(ctx, ContextKeyStepID, stepID)
	return ctx
}

// GetWorkflowFromContext extracts workflow information from the context
func GetWorkflowFromContext(ctx context.Context) (workflowID, stepID string, ok bool) {
	workflowID, wfOk := ctx.Value(ContextKeyWorkflowID).(string)
	stepID, stepOk := ctx.Value(ContextKeyStepID).(string)
	ok = (wfOk && workflowID != "") || (stepOk && stepID != "")
	return workflowID, stepID, ok
}

// WithUserIDContext adds user ID to the context
func WithUserIDContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// GetUserIDFromContext extracts user ID from the context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(ContextKeyUserID).(string)
	return userID, ok && userID != ""
}

// WithToolContext adds tool information to the context
func WithToolContext(ctx context.Context, toolName, toolSlug string) context.Context {
	ctx = context.WithValue(ctx, ContextKeyToolName, toolName)
	ctx = context.WithValue(ctx, ContextKeyToolSlug, toolSlug)
	return ctx
}

// GetToolFromContext extracts tool information from the context
func GetToolFromContext(ctx context.Context) (toolName, toolSlug string, ok bool) {
	toolName, nameOk := ctx.Value(ContextKeyToolName).(string)
	toolSlug, slugOk := ctx.Value(ContextKeyToolSlug).(string)
	ok = (nameOk && toolName != "") || (slugOk && toolSlug != "")
	return toolName, toolSlug, ok
}

// BuildAttributionFromContext extracts all available attribution information from the context
// and returns a populated AttributionContext
func BuildAttributionFromContext(ctx context.Context) *types.AttributionContext {
	attr := types.NewAttributionContext()

	// Extract agent info
	if agentID, agentName, agentType, ok := GetAgentFromContext(ctx); ok {
		attr.WithAgent(agentID, agentName, agentType)
	}

	// Extract message info
	if messageID, conversationID, ok := GetMessageFromContext(ctx); ok {
		attr.WithMessage(messageID, conversationID)
	}

	// Extract workflow info
	if workflowID, stepID, ok := GetWorkflowFromContext(ctx); ok {
		attr.WithWorkflow(workflowID, stepID)
	}

	// Extract tool info
	if toolName, toolSlug, ok := GetToolFromContext(ctx); ok {
		attr.WithTool(toolName, toolSlug)
	}

	return attr
}

// EnrichContext adds all attribution-related values from an AttributionContext to the Go context
func EnrichContext(ctx context.Context, attr *types.AttributionContext) context.Context {
	if attr == nil {
		return ctx
	}

	if attr.AgentID != "" || attr.AgentName != "" || attr.AgentType != "" {
		ctx = WithAgentContext(ctx, attr.AgentID, attr.AgentName, attr.AgentType)
	}

	if attr.MessageID != "" || attr.ConversationID != "" {
		ctx = WithMessageContext(ctx, attr.MessageID, attr.ConversationID)
	}

	if attr.WorkflowID != "" || attr.StepID != "" {
		ctx = WithWorkflowContext(ctx, attr.WorkflowID, attr.StepID)
	}

	if attr.ToolName != "" || attr.ToolSlug != "" {
		ctx = WithToolContext(ctx, attr.ToolName, attr.ToolSlug)
	}

	return ctx
}
