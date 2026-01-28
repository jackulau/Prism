package openrouter

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

const (
	defaultBaseURL = "https://openrouter.ai/api/v1"
)

// Client implements the LLM provider interface for OpenRouter
type Client struct {
	apiKey  string
	baseURL string
	siteURL string // Optional: for attribution (HTTP-Referer header)
	client  *http.Client

	// Cached models from API
	cachedModels []llm.Model
	modelsMu     sync.RWMutex
}

// NewClient creates a new OpenRouter client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		client:  &http.Client{},
	}
}

// NewClientWithOptions creates a new OpenRouter client with custom options
func NewClientWithOptions(apiKey, siteURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		siteURL: siteURL,
		client:  &http.Client{},
	}
}

// Name returns the provider name
func (c *Client) Name() string {
	return "openrouter"
}

// Models returns available models
func (c *Client) Models() []llm.Model {
	c.modelsMu.RLock()
	if len(c.cachedModels) > 0 {
		models := c.cachedModels
		c.modelsMu.RUnlock()
		return models
	}
	c.modelsMu.RUnlock()

	// Return default popular models if no cached models
	return GetPopularModels()
}

// RefreshModels fetches the latest model list from OpenRouter API
func (c *Client) RefreshModels(ctx context.Context) error {
	if c.apiKey == "" {
		return fmt.Errorf("API key not configured")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var modelsResp modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]llm.Model, 0, len(modelsResp.Data))
	for _, m := range modelsResp.Data {
		models = append(models, llm.Model{
			ID:             m.ID,
			Name:           m.Name,
			Description:    m.Description,
			ContextWindow:  m.ContextLength,
			SupportsTools:  supportsTools(m),
			SupportsVision: supportsVision(m),
		})
	}

	c.modelsMu.Lock()
	c.cachedModels = models
	c.modelsMu.Unlock()

	return nil
}

// SupportsTools returns whether the provider supports tool calling
func (c *Client) SupportsTools() bool {
	return true // Many OpenRouter models support tools
}

// SupportsVision returns whether the provider supports vision
func (c *Client) SupportsVision() bool {
	return true // Many OpenRouter models support vision
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
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("invalid API key")
		}
		return fmt.Errorf("API error: status %d", resp.StatusCode)
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
	if c.apiKey == "" {
		return nil, fmt.Errorf("OpenRouter API key not configured")
	}

	// Build request body (OpenAI-compatible format)
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

	// OpenRouter-specific headers for attribution
	if c.siteURL != "" {
		httpReq.Header.Set("HTTP-Referer", c.siteURL)
	}
	httpReq.Header.Set("X-Title", "Prism")

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
			"role": msg.Role,
		}

		// Handle messages with images (vision support)
		if len(msg.Images) > 0 {
			content := make([]map[string]interface{}, 0, len(msg.Images)+1)

			// Add text content if present
			if msg.Content != "" {
				content = append(content, map[string]interface{}{
					"type": "text",
					"text": msg.Content,
				})
			}

			// Add image content
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

		// Handle tool calls
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

		// Handle tool result messages
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

// modelsResponse represents the OpenRouter models API response
type modelsResponse struct {
	Data []modelInfo `json:"data"`
}

// modelInfo represents a model from the OpenRouter API
type modelInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	ContextLength int    `json:"context_length"`
	Architecture  struct {
		Modality string `json:"modality"`
	} `json:"architecture"`
	Pricing struct {
		Prompt     string `json:"prompt"`
		Completion string `json:"completion"`
	} `json:"pricing"`
}

// supportsTools checks if a model supports tool calling
func supportsTools(m modelInfo) bool {
	// Most modern models support tool calling
	// Check for known tool-capable models
	toolCapableModels := map[string]bool{
		"anthropic/claude-3-opus":     true,
		"anthropic/claude-3-sonnet":   true,
		"anthropic/claude-3-haiku":    true,
		"anthropic/claude-3.5-sonnet": true,
		"anthropic/claude-4-opus":     true,
		"anthropic/claude-4-sonnet":   true,
		"openai/gpt-4":                true,
		"openai/gpt-4-turbo":          true,
		"openai/gpt-4o":               true,
		"openai/gpt-4o-mini":          true,
		"google/gemini-pro":           true,
		"google/gemini-pro-1.5":       true,
		"mistralai/mistral-large":     true,
		"mistralai/mistral-medium":    true,
		"meta-llama/llama-3.1-70b":    true,
		"meta-llama/llama-3.1-405b":   true,
		"meta-llama/llama-3.3-70b":    true,
	}

	// Check exact match
	if toolCapableModels[m.ID] {
		return true
	}

	// Check prefix match for versioned models
	for prefix := range toolCapableModels {
		if strings.HasPrefix(m.ID, prefix) {
			return true
		}
	}

	return false
}

// supportsVision checks if a model supports vision/images
func supportsVision(m modelInfo) bool {
	return m.Architecture.Modality == "multimodal" ||
		strings.Contains(m.ID, "vision") ||
		strings.Contains(m.ID, "gpt-4o") ||
		strings.Contains(m.ID, "gpt-4-turbo") ||
		strings.Contains(m.ID, "claude-3") ||
		strings.Contains(m.ID, "gemini")
}
