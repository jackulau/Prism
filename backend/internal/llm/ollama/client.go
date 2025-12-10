package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jacklau/prism/internal/llm"
)

// Client implements the LLM provider interface for Ollama
type Client struct {
	baseURL string
	client  *http.Client
}

// NewClient creates a new Ollama client
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// Name returns the provider name
func (c *Client) Name() string {
	return "ollama"
}

// Models returns available models by querying Ollama
func (c *Client) Models() []llm.Model {
	// Fetch models from Ollama
	resp, err := c.client.Get(c.baseURL + "/api/tags")
	if err != nil {
		return []llm.Model{}
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name       string `json:"name"`
			ModifiedAt string `json:"modified_at"`
			Size       int64  `json:"size"`
			Details    struct {
				Format            string   `json:"format"`
				Family            string   `json:"family"`
				Families          []string `json:"families"`
				ParameterSize     string   `json:"parameter_size"`
				QuantizationLevel string   `json:"quantization_level"`
			} `json:"details"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return []llm.Model{}
	}

	models := make([]llm.Model, len(result.Models))
	for i, m := range result.Models {
		// Detect model capabilities based on model name
		supportsTools := isToolCapableModel(m.Name)
		supportsVision := isVisionCapableModel(m.Name)

		// Better context window estimation based on model
		contextWindow := getModelContextWindow(m.Name)

		models[i] = llm.Model{
			ID:             m.Name,
			Name:           m.Name,
			Description:    fmt.Sprintf("Local model - %s", m.Details.ParameterSize),
			ContextWindow:  contextWindow,
			SupportsTools:  supportsTools,
			SupportsVision: supportsVision,
		}
	}

	return models
}

// SupportsTools returns whether the provider supports tool calling
func (c *Client) SupportsTools() bool {
	return true // Ollama 0.3.0+ supports tool calling for compatible models
}

// SupportsVision returns whether the provider supports vision
func (c *Client) SupportsVision() bool {
	return false // Depends on model, but generally limited
}

// ValidateKey validates the connection (Ollama doesn't use API keys)
func (c *Client) ValidateKey(ctx context.Context, key string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama not available: status %d", resp.StatusCode)
	}

	return nil
}

// HasConfiguredKey returns true since Ollama doesn't require an API key
func (c *Client) HasConfiguredKey() bool {
	return true // Ollama is local and doesn't need an API key
}

// SetAPIKey is a no-op for Ollama since it doesn't use API keys
func (c *Client) SetAPIKey(key string) {
	// No-op: Ollama doesn't use API keys
}

// Chat sends a chat request and returns a streaming response
func (c *Client) Chat(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	// Build request body with tool call/result support
	messages := make([]map[string]interface{}, 0, len(req.Messages))
	for _, msg := range req.Messages {
		m := map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}

		// Handle assistant messages with tool calls (for conversation history)
		if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			toolCalls := make([]map[string]interface{}, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				toolCalls[j] = map[string]interface{}{
					"function": map[string]interface{}{
						"name":      tc.Name,
						"arguments": tc.Parameters,
					},
				}
			}
			m["tool_calls"] = toolCalls
		}

		// Handle tool results - Ollama expects role="tool"
		if msg.Role == "tool" && msg.ToolCallID != "" {
			m["role"] = "tool"
			// For Ollama, the content should be the tool result
		}

		messages = append(messages, m)
	}

	body := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
		"stream":   true,
	}

	// Add tools if provided and model supports them
	if len(req.Tools) > 0 && isToolCapableModel(req.Model) {
		body["tools"] = c.convertTools(req.Tools)
	}

	if req.Temperature > 0 {
		body["options"] = map[string]interface{}{
			"temperature": req.Temperature,
		}
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Create channel for streaming
	chunks := make(chan llm.StreamChunk, 100)

	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		var pendingToolCalls []llm.ToolCall

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var streamResp ollamaStreamResponse
			if err := json.Unmarshal(line, &streamResp); err != nil {
				continue
			}

			// Handle text content
			if streamResp.Message.Content != "" {
				chunks <- llm.StreamChunk{
					Delta: streamResp.Message.Content,
				}
			}

			// Handle tool calls from response
			if len(streamResp.Message.ToolCalls) > 0 {
				for i, tc := range streamResp.Message.ToolCalls {
					if tc.Function.Name != "" {
						toolCall := llm.ToolCall{
							// Generate unique ID since Ollama doesn't provide one
							ID:         fmt.Sprintf("ollama-tool-%d-%d", time.Now().UnixNano(), i),
							Name:       tc.Function.Name,
							Parameters: tc.Function.Arguments,
						}
						// Ensure Parameters is not nil
						if toolCall.Parameters == nil {
							toolCall.Parameters = make(map[string]interface{})
						}
						pendingToolCalls = append(pendingToolCalls, toolCall)
					}
				}
			}

			if streamResp.Done {
				// If we have tool calls, emit them with tool_calls finish reason
				if len(pendingToolCalls) > 0 {
					chunks <- llm.StreamChunk{
						ToolCalls:    pendingToolCalls,
						FinishReason: "tool_calls",
						Usage: &llm.Usage{
							PromptTokens:     streamResp.PromptEvalCount,
							CompletionTokens: streamResp.EvalCount,
							TotalTokens:      streamResp.PromptEvalCount + streamResp.EvalCount,
						},
					}
				} else {
					chunks <- llm.StreamChunk{
						FinishReason: "stop",
						Usage: &llm.Usage{
							PromptTokens:     streamResp.PromptEvalCount,
							CompletionTokens: streamResp.EvalCount,
							TotalTokens:      streamResp.PromptEvalCount + streamResp.EvalCount,
						},
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

// Generate sends a completion request (non-chat format)
func (c *Client) Generate(ctx context.Context, model, prompt string) (<-chan llm.StreamChunk, error) {
	body := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": true,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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

	chunks := make(chan llm.StreamChunk, 100)

	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var streamResp struct {
				Response string `json:"response"`
				Done     bool   `json:"done"`
			}
			if err := json.Unmarshal(line, &streamResp); err != nil {
				continue
			}

			if streamResp.Response != "" {
				chunks <- llm.StreamChunk{
					Delta: streamResp.Response,
				}
			}

			if streamResp.Done {
				chunks <- llm.StreamChunk{
					FinishReason: "stop",
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

// ollamaToolCall represents a tool call in Ollama response
type ollamaToolCall struct {
	Function struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	} `json:"function"`
}

// ollamaStreamResponse represents an Ollama streaming response
type ollamaStreamResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Message   struct {
		Role      string           `json:"role"`
		Content   string           `json:"content"`
		ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
	} `json:"message"`
	Done            bool  `json:"done"`
	TotalDuration   int64 `json:"total_duration"`
	LoadDuration    int64 `json:"load_duration"`
	PromptEvalCount int   `json:"prompt_eval_count"`
	EvalCount       int   `json:"eval_count"`
	EvalDuration    int64 `json:"eval_duration"`
}

// isToolCapableModel checks if a model supports tool calling
func isToolCapableModel(modelName string) bool {
	// Comprehensive list of models known to support tool/function calling
	// This list covers major open source models with native tool support
	toolCapableModels := []string{
		// Meta Llama models (3.1+ support tools)
		"llama3.1", "llama3.2", "llama3.3", "llama-3.1", "llama-3.2", "llama-3.3",

		// Mistral AI models
		"mistral", "mixtral", "codestral", "mistral-nemo", "mistral-small", "mistral-large",

		// Alibaba Qwen models
		"qwen2", "qwen2.5", "qwen-2", "qwen-2.5", "qwq",

		// Cohere Command models
		"command-r", "command-r-plus", "c4ai",

		// Function-calling specialized models
		"firefunction", "functionary", "gorilla", "nexusraven",

		// Nous Research models
		"hermes", "nous-hermes", "openhermes",

		// IBM Granite models
		"granite",

		// DeepSeek models
		"deepseek", "deepseek-coder", "deepseek-v2", "deepseek-v3",

		// Microsoft Phi models
		"phi-3", "phi-4", "phi3", "phi4",

		// Google Gemma models (2.0+ with tool support)
		"gemma2", "gemma-2",

		// 01.ai Yi models
		"yi-", "yi1.5", "yi-1.5",

		// Other tool-capable models
		"internlm", "glm", "chatglm",
		"solar", "solar-pro",
		"dolphin", "dolphin-mistral", "dolphin-llama",
		"nemotron",
		"smollm2",
		"athene",
		"marco",
	}

	nameLower := strings.ToLower(modelName)
	for _, tcm := range toolCapableModels {
		if strings.Contains(nameLower, tcm) {
			return true
		}
	}
	return false
}

// isVisionCapableModel checks if a model supports vision/image input
func isVisionCapableModel(modelName string) bool {
	visionModels := []string{
		// LLaVA models
		"llava", "llava-llama3", "llava-phi3",
		// Generic vision indicators
		"vision", "-v", "vl",
		// Specific vision models
		"bakllava", "moondream",
		"minicpm-v", "internvl",
		"cogvlm", "qwen-vl", "qwen2-vl",
		"llama3.2-vision", "llama-3.2-vision",
		"pixtral",
	}

	nameLower := strings.ToLower(modelName)
	for _, vm := range visionModels {
		if strings.Contains(nameLower, vm) {
			return true
		}
	}
	return false
}

// convertTools converts llm.ToolDefinition to Ollama format (OpenAI-compatible)
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

// getModelContextWindow returns the estimated context window for a model
func getModelContextWindow(modelName string) int {
	nameLower := strings.ToLower(modelName)

	// Models with 128k+ context
	if strings.Contains(nameLower, "llama3.1") || strings.Contains(nameLower, "llama-3.1") ||
		strings.Contains(nameLower, "llama3.2") || strings.Contains(nameLower, "llama-3.2") ||
		strings.Contains(nameLower, "llama3.3") || strings.Contains(nameLower, "llama-3.3") {
		return 131072 // 128k
	}
	if strings.Contains(nameLower, "qwen2.5") || strings.Contains(nameLower, "qwen-2.5") ||
		strings.Contains(nameLower, "qwq") {
		return 131072 // 128k
	}
	if strings.Contains(nameLower, "deepseek-v3") || strings.Contains(nameLower, "deepseek-v2") {
		return 131072 // 128k
	}
	if strings.Contains(nameLower, "mistral-large") || strings.Contains(nameLower, "mistral-nemo") {
		return 131072 // 128k
	}
	if strings.Contains(nameLower, "gemma2") || strings.Contains(nameLower, "gemma-2") {
		return 8192
	}

	// Models with 32k context
	if strings.Contains(nameLower, "mistral") || strings.Contains(nameLower, "mixtral") {
		return 32768 // 32k
	}
	if strings.Contains(nameLower, "qwen2") || strings.Contains(nameLower, "qwen-2") {
		return 32768 // 32k
	}
	if strings.Contains(nameLower, "command-r") {
		return 131072 // 128k
	}
	if strings.Contains(nameLower, "yi-") || strings.Contains(nameLower, "yi1.5") {
		return 32768 // 32k
	}
	if strings.Contains(nameLower, "deepseek") {
		return 32768 // 32k
	}

	// Models with 16k context
	if strings.Contains(nameLower, "phi-3") || strings.Contains(nameLower, "phi3") ||
		strings.Contains(nameLower, "phi-4") || strings.Contains(nameLower, "phi4") {
		return 16384 // 16k
	}
	if strings.Contains(nameLower, "granite") {
		return 8192
	}

	// Models with 8k context
	if strings.Contains(nameLower, "llama3") || strings.Contains(nameLower, "llama-3") {
		return 8192 // 8k (base llama3 without extended context)
	}
	if strings.Contains(nameLower, "hermes") || strings.Contains(nameLower, "openhermes") {
		return 8192
	}

	// Default context window
	return 4096
}
