package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jacklau/prism/internal/llm"
)

// Client implements the LLM provider interface for Anthropic
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewClient creates a new Anthropic client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1",
		client:  &http.Client{},
	}
}

// Name returns the provider name
func (c *Client) Name() string {
	return "anthropic"
}

// Models returns available models
func (c *Client) Models() []llm.Model {
	return []llm.Model{
		{
			ID:             "claude-3-5-sonnet-20241022",
			Name:           "Claude 3.5 Sonnet",
			Description:    "Most intelligent model, best for complex tasks",
			ContextWindow:  200000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "claude-3-5-haiku-20241022",
			Name:           "Claude 3.5 Haiku",
			Description:    "Fastest and most cost-effective",
			ContextWindow:  200000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "claude-3-opus-20240229",
			Name:           "Claude 3 Opus",
			Description:    "Powerful model for complex analysis",
			ContextWindow:  200000,
			SupportsTools:  true,
			SupportsVision: true,
		},
	}
}

// SupportsTools returns whether the provider supports tool calling
func (c *Client) SupportsTools() bool {
	return true
}

// SupportsVision returns whether the provider supports vision
func (c *Client) SupportsVision() bool {
	return true
}

// ValidateKey validates an API key
func (c *Client) ValidateKey(ctx context.Context, key string) error {
	// Anthropic doesn't have a simple key validation endpoint
	// We'll try a minimal request
	body := map[string]interface{}{
		"model":      "claude-3-5-haiku-20241022",
		"max_tokens": 1,
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hi"},
		},
	}

	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("x-api-key", key)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}

	return nil
}

// Chat sends a chat request and returns a streaming response
func (c *Client) Chat(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	// Extract system message
	var systemMessage string
	messages := make([]map[string]interface{}, 0)

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemMessage = msg.Content
			continue
		}

		m := map[string]interface{}{
			"role": msg.Role,
		}

		// Handle content with images
		if len(msg.Images) > 0 {
			content := []map[string]interface{}{}
			for _, img := range msg.Images {
				if img.Base64 != "" {
					content = append(content, map[string]interface{}{
						"type": "image",
						"source": map[string]interface{}{
							"type":         "base64",
							"media_type":   img.MimeType,
							"data":         img.Base64,
						},
					})
				}
			}
			if msg.Content != "" {
				content = append(content, map[string]interface{}{
					"type": "text",
					"text": msg.Content,
				})
			}
			m["content"] = content
		} else {
			m["content"] = msg.Content
		}

		// Handle tool results
		if msg.ToolCallID != "" {
			m["role"] = "user"
			m["content"] = []map[string]interface{}{
				{
					"type":        "tool_result",
					"tool_use_id": msg.ToolCallID,
					"content":     msg.Content,
				},
			}
		}

		messages = append(messages, m)
	}

	// Build request body
	body := map[string]interface{}{
		"model":      req.Model,
		"messages":   messages,
		"max_tokens": 4096,
		"stream":     true,
	}

	if systemMessage != "" {
		body["system"] = systemMessage
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if len(req.Tools) > 0 {
		body["tools"] = c.convertTools(req.Tools)
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Create channel for streaming
	chunks := make(chan llm.StreamChunk, 100)

	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		var currentToolCall *llm.ToolCall
		var toolInputJSON strings.Builder

		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines
			if line == "" {
				continue
			}

			// Handle event type
			if strings.HasPrefix(line, "event: ") {
				continue
			}

			// Handle data
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			var event streamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			switch event.Type {
			case "content_block_start":
				if event.ContentBlock.Type == "tool_use" {
					currentToolCall = &llm.ToolCall{
						ID:         event.ContentBlock.ID,
						Name:       event.ContentBlock.Name,
						Parameters: make(map[string]interface{}),
					}
					toolInputJSON.Reset()
				}

			case "content_block_delta":
				if event.Delta.Type == "text_delta" {
					chunks <- llm.StreamChunk{
						Delta: event.Delta.Text,
					}
				} else if event.Delta.Type == "input_json_delta" {
					toolInputJSON.WriteString(event.Delta.PartialJSON)
				}

			case "content_block_stop":
				if currentToolCall != nil {
					// Parse complete tool input
					if toolInputJSON.Len() > 0 {
						json.Unmarshal([]byte(toolInputJSON.String()), &currentToolCall.Parameters)
					}
					chunks <- llm.StreamChunk{
						ToolCalls: []llm.ToolCall{*currentToolCall},
					}
					currentToolCall = nil
				}

			case "message_stop":
				chunks <- llm.StreamChunk{
					FinishReason: "stop",
				}

			case "message_delta":
				if event.Delta.StopReason != "" {
					chunks <- llm.StreamChunk{
						FinishReason: event.Delta.StopReason,
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			chunks <- llm.StreamChunk{
				Error: err,
			}
		}
	}()

	return chunks, nil
}

// convertTools converts llm.ToolDefinition to Anthropic format
func (c *Client) convertTools(tools []llm.ToolDefinition) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tools))

	for i, tool := range tools {
		result[i] = map[string]interface{}{
			"name":         tool.Name,
			"description":  tool.Description,
			"input_schema": tool.Parameters,
		}
	}

	return result
}

// streamEvent represents an Anthropic streaming event
type streamEvent struct {
	Type         string `json:"type"`
	Index        int    `json:"index"`
	ContentBlock struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Name  string `json:"name"`
		Input struct{}
	} `json:"content_block"`
	Delta struct {
		Type        string `json:"type"`
		Text        string `json:"text"`
		PartialJSON string `json:"partial_json"`
		StopReason  string `json:"stop_reason"`
	} `json:"delta"`
}
