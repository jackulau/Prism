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

// ClientOption configures the Client
type ClientOption func(*Client)

// Client implements the LLM provider interface for OpenAI
type Client struct {
	apiKey       string
	baseURL      string
	client       *http.Client
	tokenCounter *TokenCounter
	usageTracker *UsageTracker
	rateLimiter  *RateLimiter
	retryConfig  *RetryConfig
}

// WithBaseURL sets a custom base URL
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.client = client
	}
}

// WithRateLimiter sets a custom rate limiter
func WithRateLimiter(rl *RateLimiter) ClientOption {
	return func(c *Client) {
		c.rateLimiter = rl
	}
}

// WithRetryConfig sets a custom retry configuration
func WithRetryConfig(cfg *RetryConfig) ClientOption {
	return func(c *Client) {
		c.retryConfig = cfg
	}
}

// WithUsageTracking enables usage tracking
func WithUsageTracking() ClientOption {
	return func(c *Client) {
		c.usageTracker = NewUsageTracker()
	}
}

// NewClient creates a new OpenAI client
func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:      apiKey,
		baseURL:     "https://api.openai.com/v1",
		client:      &http.Client{},
		retryConfig: DefaultRetryConfig(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Name returns the provider name
func (c *Client) Name() string {
	return "openai"
}

// Models returns available models
func (c *Client) Models() []llm.Model {
	models := make([]llm.Model, 0, len(ModelConfigs))
	for _, config := range ModelConfigs {
		models = append(models, llm.Model{
			ID:             config.ID,
			Name:           config.Name,
			ContextWindow:  config.ContextWindow,
			SupportsTools:  config.SupportsTools,
			SupportsVision: config.SupportsVision,
		})
	}
	return models
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

// GetUsageTracker returns the usage tracker (may be nil if not enabled)
func (c *Client) GetUsageTracker() *UsageTracker {
	return c.usageTracker
}

// GetTokenCounter returns a token counter for the specified model
func (c *Client) GetTokenCounter(model string) (*TokenCounter, error) {
	return NewTokenCounter(model)
}

// Chat sends a chat request and returns a streaming response
func (c *Client) Chat(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	return c.ChatWithOptions(ctx, req, nil)
}

// ChatWithOptions sends a chat request with additional options
func (c *Client) ChatWithOptions(ctx context.Context, req *llm.ChatRequest, opts *ChatOptions) (<-chan llm.StreamChunk, error) {
	// Validate options if provided
	if opts != nil {
		if err := opts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options: %w", err)
		}
	}

	// Wait for rate limiter if configured
	if c.rateLimiter != nil {
		// Estimate tokens for rate limiting
		estimatedTokens := EstimateMessagesTokens(req.Messages)
		if err := c.rateLimiter.Wait(ctx, estimatedTokens); err != nil {
			return nil, fmt.Errorf("rate limit wait failed: %w", err)
		}
	}

	// Build request body
	body := map[string]interface{}{
		"model":  req.Model,
		"stream": true,
	}

	// Convert messages with vision support if needed
	if hasImages(req.Messages) {
		if err := ValidateVisionSupport(req.Model); err != nil {
			return nil, err
		}
		messages, err := c.convertMessagesWithVision(req.Messages)
		if err != nil {
			return nil, fmt.Errorf("failed to convert messages: %w", err)
		}
		body["messages"] = messages
	} else {
		body["messages"] = c.convertMessages(req.Messages)
	}

	// Apply basic request parameters
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}
	if len(req.Tools) > 0 {
		body["tools"] = c.convertTools(req.Tools)
	}

	// Apply options if provided
	if opts != nil {
		for k, v := range opts.ToRequestBody() {
			body[k] = v
		}
	}

	// Include stream options to get usage in response
	body["stream_options"] = map[string]interface{}{
		"include_usage": true,
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

	// Execute request with retry logic if configured
	var resp *http.Response
	if c.retryConfig != nil && c.retryConfig.MaxRetries > 0 {
		retryable := NewRetryableRequest(c.retryConfig, c.rateLimiter)
		resp, err = retryable.Do(ctx, c.client, httpReq)
	} else {
		resp, err = c.client.Do(httpReq)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Update rate limiter from response headers
	if c.rateLimiter != nil {
		c.rateLimiter.UpdateFromHeaders(resp.Header)
	}

	// Create channel for streaming
	chunks := make(chan llm.StreamChunk, 100)

	go c.handleStreamResponse(resp, chunks, req.Model)

	return chunks, nil
}

// handleStreamResponse processes the streaming response
func (c *Client) handleStreamResponse(resp *http.Response, chunks chan<- llm.StreamChunk, model string) {
	defer close(chunks)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)

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

		// Handle usage statistics (comes in final chunk)
		if streamResp.Usage != nil && c.usageTracker != nil {
			usage := &llm.Usage{
				PromptTokens:     streamResp.Usage.PromptTokens,
				CompletionTokens: streamResp.Usage.CompletionTokens,
				TotalTokens:      streamResp.Usage.TotalTokens,
			}
			c.usageTracker.Record(model, usage)
			chunks <- llm.StreamChunk{
				Usage: usage,
			}
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
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}
