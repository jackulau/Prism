package lmstudio

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

// Client implements the LLM provider interface for LM Studio
type Client struct {
	baseURL string
	client  *http.Client
}

// NewClient creates a new LM Studio client
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:1234"
	}
	// Ensure we don't double the /v1 suffix
	baseURL = strings.TrimSuffix(baseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/v1")

	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// Name returns the provider name
func (c *Client) Name() string {
	return "lmstudio"
}

// Models returns available models by querying LM Studio
func (c *Client) Models() []llm.Model {
	// Try to fetch loaded models from LM Studio
	models, err := c.fetchModels()
	if err != nil {
		return []llm.Model{}
	}
	return models
}

// SupportsTools returns whether the provider supports tool calling
func (c *Client) SupportsTools() bool {
	return true // LM Studio supports tools via OpenAI-compatible API
}

// SupportsVision returns whether the provider supports vision
func (c *Client) SupportsVision() bool {
	return false // Vision support depends on loaded model
}

// ValidateKey validates the connection (LM Studio doesn't use API keys)
func (c *Client) ValidateKey(ctx context.Context, key string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/v1/models", nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to LM Studio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("LM Studio not available: status %d", resp.StatusCode)
	}

	return nil
}

// HasConfiguredKey returns true since LM Studio doesn't require an API key
func (c *Client) HasConfiguredKey() bool {
	return true // LM Studio is local and doesn't need an API key
}

// SetAPIKey is a no-op for LM Studio since it doesn't use API keys
func (c *Client) SetAPIKey(key string) {
	// No-op: LM Studio doesn't use API keys
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

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to LM Studio: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("LM Studio API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Create channel for streaming
	chunks := make(chan llm.StreamChunk, 100)

	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)

		// Accumulator for tool calls - we need to accumulate arguments as strings
		// before parsing, since they come in partial JSON chunks
		type toolCallAccumulator struct {
			ID        string
			Name      string
			Arguments strings.Builder
		}
		var toolCallAccumulators []toolCallAccumulator

		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and non-data lines
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// Check for stream end
			if data == "[DONE]" {
				// Parse accumulated arguments and emit final tool calls
				if len(toolCallAccumulators) > 0 {
					finalToolCalls := make([]llm.ToolCall, len(toolCallAccumulators))
					for i, acc := range toolCallAccumulators {
						var params map[string]interface{}
						argsStr := acc.Arguments.String()
						if argsStr != "" {
							if err := json.Unmarshal([]byte(argsStr), &params); err != nil {
								params = make(map[string]interface{})
							}
						} else {
							params = make(map[string]interface{})
						}
						finalToolCalls[i] = llm.ToolCall{
							ID:         acc.ID,
							Name:       acc.Name,
							Parameters: params,
						}
					}
					chunks <- llm.StreamChunk{
						ToolCalls:    finalToolCalls,
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

			// Handle tool calls - accumulate arguments as strings
			if len(choice.Delta.ToolCalls) > 0 {
				for _, tc := range choice.Delta.ToolCalls {
					// Extend accumulators if needed
					for tc.Index >= len(toolCallAccumulators) {
						toolCallAccumulators = append(toolCallAccumulators, toolCallAccumulator{})
					}

					// Set ID and Name if provided (they come in the first chunk)
					if tc.ID != "" {
						toolCallAccumulators[tc.Index].ID = tc.ID
					}
					if tc.Function.Name != "" {
						toolCallAccumulators[tc.Index].Name = tc.Function.Name
					}

					// Accumulate argument chunks (partial JSON strings)
					if tc.Function.Arguments != "" {
						toolCallAccumulators[tc.Index].Arguments.WriteString(tc.Function.Arguments)
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

// streamResponse represents an OpenAI-compatible streaming response
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
