package anthropic

import (
	"time"

	"github.com/jacklau/prism/internal/cloudprovider"
)

// API request/response types for Anthropic's agent API

// createAgentRequest represents the request to create a new agent
type createAgentRequest struct {
	Model        string                  `json:"model"`
	Name         string                  `json:"name,omitempty"`
	SystemPrompt string                  `json:"system,omitempty"`
	Tools        []toolDefinitionRequest `json:"tools,omitempty"`
	Metadata     map[string]interface{}  `json:"metadata,omitempty"`
}

// toolDefinitionRequest represents a tool definition for the API
type toolDefinitionRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
}

// agentResponse represents the response from agent endpoints
type agentResponse struct {
	ID        string                 `json:"id"`
	Object    string                 `json:"object"`
	Model     string                 `json:"model"`
	Name      string                 `json:"name,omitempty"`
	CreatedAt int64                  `json:"created_at"`
	Status    string                 `json:"status,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// sendMessageRequest represents the request to send a message
type sendMessageRequest struct {
	Model     string                 `json:"model"`
	MaxTokens int                    `json:"max_tokens"`
	Messages  []messageRequest       `json:"messages"`
	System    string                 `json:"system,omitempty"`
	Stream    bool                   `json:"stream,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// messageRequest represents a message in the API request
type messageRequest struct {
	Role    string        `json:"role"`
	Content []contentPart `json:"content"`
}

// contentPart represents a content part in a message
type contentPart struct {
	Type      string       `json:"type"`
	Text      string       `json:"text,omitempty"`
	Source    *imageSource `json:"source,omitempty"`
	ID        string       `json:"id,omitempty"`
	Name      string       `json:"name,omitempty"`
	Input     interface{}  `json:"input,omitempty"`
	ToolUseID string       `json:"tool_use_id,omitempty"`
	Content   string       `json:"content,omitempty"`
}

// imageSource represents an image source
type imageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// messageResponse represents the response from message endpoints
type messageResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []contentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason,omitempty"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        *usageInfo     `json:"usage,omitempty"`
}

// contentBlock represents a content block in the response
type contentBlock struct {
	Type  string                 `json:"type"`
	Text  string                 `json:"text,omitempty"`
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}

// usageInfo represents token usage information
type usageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// streamEvent represents a streaming event from the API
type streamEvent struct {
	Type         string           `json:"type"`
	Index        int              `json:"index"`
	ContentBlock *contentBlock    `json:"content_block,omitempty"`
	Delta        *streamDelta     `json:"delta,omitempty"`
	Message      *messageResponse `json:"message,omitempty"`
	Usage        *usageInfo       `json:"usage,omitempty"`
}

// streamDelta represents a delta in a streaming response
type streamDelta struct {
	Type        string `json:"type"`
	Text        string `json:"text,omitempty"`
	PartialJSON string `json:"partial_json,omitempty"`
	StopReason  string `json:"stop_reason,omitempty"`
}

// errorResponse represents an API error response
type errorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Conversion helpers

// toAgent converts an agentResponse to a cloudprovider.Agent
func (r *agentResponse) toAgent() *cloudprovider.Agent {
	status := cloudprovider.AgentStatusActive
	if r.Status != "" {
		switch r.Status {
		case "idle":
			status = cloudprovider.AgentStatusIdle
		case "processing":
			status = cloudprovider.AgentStatusProcessing
		case "terminated":
			status = cloudprovider.AgentStatusTerminated
		}
	}

	return &cloudprovider.Agent{
		ID:         r.ID,
		ProviderID: r.ID,
		Name:       r.Name,
		Model:      r.Model,
		Status:     status,
		CreatedAt:  time.Unix(r.CreatedAt, 0),
		UpdatedAt:  time.Unix(r.CreatedAt, 0),
		Metadata:   r.Metadata,
	}
}

// toProviderMessages converts a messageResponse to cloudprovider.ProviderMessage
func (r *messageResponse) toProviderMessages() []cloudprovider.ProviderMessage {
	var messages []cloudprovider.ProviderMessage

	var content string
	var toolCalls []cloudprovider.ToolCall

	for _, block := range r.Content {
		switch block.Type {
		case "text":
			content += block.Text
		case "tool_use":
			toolCalls = append(toolCalls, cloudprovider.ToolCall{
				ID:         block.ID,
				Name:       block.Name,
				Parameters: block.Input,
			})
		}
	}

	msg := cloudprovider.ProviderMessage{
		ID:        r.ID,
		Role:      r.Role,
		Content:   content,
		ToolCalls: toolCalls,
		Timestamp: time.Now(),
	}

	messages = append(messages, msg)
	return messages
}

// fromCreateParams creates a createAgentRequest from CreateAgentParams
func fromCreateParams(params cloudprovider.CreateAgentParams) *createAgentRequest {
	req := &createAgentRequest{
		Model:        params.Model,
		Name:         params.Name,
		SystemPrompt: params.SystemPrompt,
		Metadata:     params.Metadata,
	}

	for _, tool := range params.Tools {
		req.Tools = append(req.Tools, toolDefinitionRequest{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.Parameters,
		})
	}

	return req
}

// fromImages converts cloudprovider.ImageData to API format
func fromImages(images []cloudprovider.ImageData) []contentPart {
	var parts []contentPart
	for _, img := range images {
		if img.Base64 != "" {
			parts = append(parts, contentPart{
				Type: "image",
				Source: &imageSource{
					Type:      "base64",
					MediaType: img.MimeType,
					Data:      img.Base64,
				},
			})
		}
	}
	return parts
}
