package openai

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

// Client implements the LLM provider interface for OpenAI
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewClient creates a new OpenAI client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		client:  &http.Client{},
	}
}

// Name returns the provider name
func (c *Client) Name() string {
	return "openai"
}

// Models returns available models
func (c *Client) Models() []llm.Model {
	return []llm.Model{
		{
			ID:             "o3",
			Name:           "OpenAI o3",
			Description:    "Most powerful reasoning model",
			ContextWindow:  200000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "o4-mini",
			Name:           "OpenAI o4-mini",
			Description:    "Fast, cost-efficient reasoning",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "gpt-4.1",
			Name:           "GPT-4.1",
			Description:    "Latest flagship model",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "gpt-4.1-mini",
			Name:           "GPT-4.1 Mini",
			Description:    "Fast and affordable",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "gpt-4o",
			Name:           "GPT-4o",
			Description:    "Legacy multimodal model",
			ContextWindow:  128000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "gpt-4o-mini",
			Name:           "GPT-4o Mini",
			Description:    "Legacy fast model",
			ContextWindow:  128000,
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
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/models", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid API key: status %d", resp.StatusCode)
	}

	return nil
}

// HasConfiguredKey returns whether the provider has an API key configured
func (c *Client) HasConfiguredKey() bool {
	return c.apiKey != ""
}

// SetAPIKey updates the provider's API key
func (c *Client) SetAPIKey(key string) {
	c.apiKey = key
}

// Chat sends a chat request and returns a streaming response
func (c *Client) Chat(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	// Build request body
	body := map[string]interface{}{
		"model":    req.Model,
		"messages": c.convertMessages(req.Messages),
		"stream":   true,
	}

	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}
	if len(req.Tools) > 0 {
		body["tools"] = c.convertTools(req.Tools)
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
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
		var currentToolCalls []llm.ToolCall

		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and non-data lines
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// Check for stream end
			if data == "[DONE]" {
				if len(currentToolCalls) > 0 {
					chunks <- llm.StreamChunk{
						ToolCalls:    currentToolCalls,
						FinishReason: "tool_calls",
					}
				}
				break
			}

			// Parse the SSE data
			var streamResp streamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue
			}

			if len(streamResp.Choices) == 0 {
				continue
			}

			choice := streamResp.Choices[0]

			// Handle content delta
			if choice.Delta.Content != "" {
				chunks <- llm.StreamChunk{
					Delta: choice.Delta.Content,
				}
			}

			// Handle tool calls
			if len(choice.Delta.ToolCalls) > 0 {
				for _, tc := range choice.Delta.ToolCalls {
					// Find or create tool call
					if tc.Index >= len(currentToolCalls) {
						currentToolCalls = append(currentToolCalls, llm.ToolCall{
							ID:   tc.ID,
							Name: tc.Function.Name,
						})
					}
					// Append arguments
					if tc.Function.Arguments != "" {
						idx := tc.Index
						if currentToolCalls[idx].Parameters == nil {
							currentToolCalls[idx].Parameters = make(map[string]interface{})
						}
						// Parse accumulated arguments
						var args map[string]interface{}
						if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err == nil {
							for k, v := range args {
								currentToolCalls[idx].Parameters[k] = v
							}
						}
					}
				}
			}

			// Check for finish reason
			if choice.FinishReason != "" {
				chunks <- llm.StreamChunk{
					FinishReason: choice.FinishReason,
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

// convertMessages converts llm.Message to OpenAI format
func (c *Client) convertMessages(messages []llm.Message) []map[string]interface{} {
	result := make([]map[string]interface{}, len(messages))

	for i, msg := range messages {
		m := map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}

		if len(msg.ToolCalls) > 0 {
			toolCalls := make([]map[string]interface{}, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				args, _ := json.Marshal(tc.Parameters)
				toolCalls[j] = map[string]interface{}{
					"id":   tc.ID,
					"type": "function",
					"function": map[string]interface{}{
						"name":      tc.Name,
						"arguments": string(args),
					},
				}
			}
			m["tool_calls"] = toolCalls
		}

		if msg.ToolCallID != "" {
			m["tool_call_id"] = msg.ToolCallID
		}

		result[i] = m
	}

	return result
}

// convertTools converts llm.ToolDefinition to OpenAI format
func (c *Client) convertTools(tools []llm.ToolDefinition) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tools))

	for i, tool := range tools {
		result[i] = map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		}
	}

	return result
}

// streamResponse represents an OpenAI streaming response
type streamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}
