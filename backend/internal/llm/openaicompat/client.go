package openaicompat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/jacklau/prism/internal/llm"
)

// Client implements the LLM provider interface for OpenAI-compatible endpoints.
// This includes vLLM, llama.cpp server, LocalAI, text-generation-webui, Jan,
// and other services that implement the OpenAI API specification.
type Client struct {
	id            string // Unique identifier for this custom provider
	name          string // User-friendly name
	apiKey        string
	baseURL       string
	client        *http.Client
	supportsTools bool
	supportsVision bool
	models        []llm.Model
	mu            sync.RWMutex
}

// Config holds configuration for a custom OpenAI-compatible provider
type Config struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	BaseURL       string      `json:"base_url"`
	APIKey        string      `json:"api_key,omitempty"`
	SupportsTools bool        `json:"supports_tools"`
	SupportsVision bool       `json:"supports_vision"`
	Models        []llm.Model `json:"models,omitempty"`
}

// NewClient creates a new OpenAI-compatible client with the given configuration
func NewClient(cfg Config) *Client {
	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")

	return &Client{
		id:            cfg.ID,
		name:          cfg.Name,
		apiKey:        cfg.APIKey,
		baseURL:       baseURL,
		client:        &http.Client{},
		supportsTools: cfg.SupportsTools,
		supportsVision: cfg.SupportsVision,
		models:        cfg.Models,
	}
}

// ID returns the unique identifier for this provider
func (c *Client) ID() string {
	return c.id
}

// Name returns the provider name (unique identifier for the LLM manager)
func (c *Client) Name() string {
	// Use the ID as the provider name since it needs to be unique
	return "custom:" + c.id
}

// DisplayName returns the user-friendly display name
func (c *Client) DisplayName() string {
	return c.name
}

// BaseURL returns the configured base URL
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Models returns available models
func (c *Client) Models() []llm.Model {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.models) > 0 {
		return c.models
	}

	// Return empty list if no models configured
	return []llm.Model{}
}

// SetModels updates the available models
func (c *Client) SetModels(models []llm.Model) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.models = models
}

// SupportsTools returns whether the provider supports tool calling
func (c *Client) SupportsTools() bool {
	return c.supportsTools
}

// SetSupportsTools updates the tool support flag
func (c *Client) SetSupportsTools(supports bool) {
	c.supportsTools = supports
}

// SupportsVision returns whether the provider supports vision
func (c *Client) SupportsVision() bool {
	return c.supportsVision
}

// SetSupportsVision updates the vision support flag
func (c *Client) SetSupportsVision(supports bool) {
	c.supportsVision = supports
}

// ValidateKey validates an API key by testing the /v1/models endpoint
func (c *Client) ValidateKey(ctx context.Context, key string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/v1/models", nil)
	if err != nil {
		return err
	}

	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid API key: status %d", resp.StatusCode)
	}

	// Accept 200 OK or 404 (models endpoint might not exist)
	// Some endpoints don't implement /v1/models
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("endpoint error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// HasConfiguredKey returns whether the provider has an API key configured
// Returns true if no API key is required (local endpoints)
func (c *Client) HasConfiguredKey() bool {
	// For OpenAI-compatible endpoints, we consider it "configured" if:
	// 1. An API key is set, OR
	// 2. The endpoint doesn't require authentication (local servers)
	return true
}

// SetAPIKey updates the provider's API key
func (c *Client) SetAPIKey(key string) {
	c.apiKey = key
}

// TestConnection tests the connection to the endpoint
func (c *Client) TestConnection(ctx context.Context) error {
	return c.ValidateKey(ctx, c.apiKey)
}

// FetchModels fetches available models from the /v1/models endpoint
func (c *Client) FetchModels(ctx context.Context) ([]llm.Model, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/v1/models", nil)
	if err != nil {
		return nil, err
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch models: status %d, body: %s", resp.StatusCode, string(body))
	}

	var modelsResp modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to parse models response: %w", err)
	}

	models := make([]llm.Model, len(modelsResp.Data))
	for i, m := range modelsResp.Data {
		models[i] = llm.Model{
			ID:             m.ID,
			Name:           m.ID, // Use ID as name if no other info available
			Description:    fmt.Sprintf("Model from %s", c.name),
			ContextWindow:  getContextWindow(m.ID),
			SupportsTools:  c.supportsTools,
			SupportsVision: c.supportsVision,
		}
	}

	return models, nil
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
	if len(req.Tools) > 0 && c.supportsTools {
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

	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
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
		// Increase buffer size for large responses
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		// Accumulator for tool calls
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

					// Set ID and Name if provided
					if tc.ID != "" {
						toolCallAccumulators[tc.Index].ID = tc.ID
					}
					if tc.Function.Name != "" {
						toolCallAccumulators[tc.Index].Name = tc.Function.Name
					}

					// Accumulate argument chunks
					if tc.Function.Arguments != "" {
						toolCallAccumulators[tc.Index].Arguments.WriteString(tc.Function.Arguments)
					}
				}
			}

			// Check for finish reason
			if choice.FinishReason != "" {
				if choice.FinishReason != "tool_calls" || len(toolCallAccumulators) == 0 {
					chunks <- llm.StreamChunk{
						FinishReason: choice.FinishReason,
					}
				}
			}

			// Handle usage if provided
			if streamResp.Usage != nil {
				chunks <- llm.StreamChunk{
					Usage: &llm.Usage{
						PromptTokens:     streamResp.Usage.PromptTokens,
						CompletionTokens: streamResp.Usage.CompletionTokens,
						TotalTokens:      streamResp.Usage.TotalTokens,
					},
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
			"role": msg.Role,
		}

		// Handle vision/images
		if len(msg.Images) > 0 && c.supportsVision {
			content := make([]map[string]interface{}, 0)

			// Add text content if present
			if msg.Content != "" {
				content = append(content, map[string]interface{}{
					"type": "text",
					"text": msg.Content,
				})
			}

			// Add images
			for _, img := range msg.Images {
				imageContent := map[string]interface{}{
					"type": "image_url",
				}
				if img.URL != "" {
					imageContent["image_url"] = map[string]interface{}{
						"url": img.URL,
					}
				} else if img.Base64 != "" {
					mimeType := img.MimeType
					if mimeType == "" {
						mimeType = "image/png"
					}
					imageContent["image_url"] = map[string]interface{}{
						"url": fmt.Sprintf("data:%s;base64,%s", mimeType, img.Base64),
					}
				}
				content = append(content, imageContent)
			}

			m["content"] = content
		} else {
			m["content"] = msg.Content
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

// getContextWindow estimates context window based on model name
func getContextWindow(modelID string) int {
	modelLower := strings.ToLower(modelID)

	// Check for common patterns
	if strings.Contains(modelLower, "128k") {
		return 128000
	}
	if strings.Contains(modelLower, "64k") {
		return 64000
	}
	if strings.Contains(modelLower, "32k") {
		return 32000
	}
	if strings.Contains(modelLower, "16k") {
		return 16000
	}
	if strings.Contains(modelLower, "8k") {
		return 8000
	}

	// Check for known model patterns
	if strings.Contains(modelLower, "llama-3") || strings.Contains(modelLower, "llama3") {
		return 8192
	}
	if strings.Contains(modelLower, "mistral") {
		return 32000
	}
	if strings.Contains(modelLower, "mixtral") {
		return 32000
	}
	if strings.Contains(modelLower, "qwen") {
		return 32000
	}
	if strings.Contains(modelLower, "phi") {
		return 4096
	}
	if strings.Contains(modelLower, "gemma") {
		return 8192
	}

	// Default context window
	return 4096
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
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// modelsResponse represents the /v1/models response
type modelsResponse struct {
	Object string `json:"object"`
	Data   []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}
