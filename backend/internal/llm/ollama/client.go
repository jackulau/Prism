package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
		models[i] = llm.Model{
			ID:             m.Name,
			Name:           m.Name,
			Description:    fmt.Sprintf("Local model - %s", m.Details.ParameterSize),
			ContextWindow:  4096, // Default, may vary by model
			SupportsTools:  false, // Most Ollama models don't support tools
			SupportsVision: false,
		}
	}

	return models
}

// SupportsTools returns whether the provider supports tool calling
func (c *Client) SupportsTools() bool {
	return false // Most Ollama models don't support tool calling yet
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

// Chat sends a chat request and returns a streaming response
func (c *Client) Chat(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	// Build request body
	messages := make([]map[string]interface{}, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	body := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
		"stream":   true,
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

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var streamResp ollamaStreamResponse
			if err := json.Unmarshal(line, &streamResp); err != nil {
				continue
			}

			if streamResp.Message.Content != "" {
				chunks <- llm.StreamChunk{
					Delta: streamResp.Message.Content,
				}
			}

			if streamResp.Done {
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

// ollamaStreamResponse represents an Ollama streaming response
type ollamaStreamResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Message   struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done            bool   `json:"done"`
	TotalDuration   int64  `json:"total_duration"`
	LoadDuration    int64  `json:"load_duration"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
	EvalDuration    int64  `json:"eval_duration"`
}
