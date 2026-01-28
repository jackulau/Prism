package perplexity

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

// Client implements the LLM provider interface for Perplexity AI
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewClient creates a new Perplexity client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.perplexity.ai",
		client:  &http.Client{},
	}
}

// Name returns the provider name
func (c *Client) Name() string {
	return "perplexity"
}

// Models returns available models
func (c *Client) Models() []llm.Model {
	return models
}

// SupportsTools returns whether the provider supports tool calling
func (c *Client) SupportsTools() bool {
	return false // Perplexity doesn't support function calling
}

// SupportsVision returns whether the provider supports vision
func (c *Client) SupportsVision() bool {
	return false // Perplexity Sonar models don't support vision
}

// ValidateKey validates an API key
func (c *Client) ValidateKey(ctx context.Context, key string) error {
	// Test the API key with a simple request
	body := map[string]interface{}{
		"model": "llama-3.1-sonar-small-128k-online",
		"messages": []map[string]string{
			{"role": "user", "content": "hi"},
		},
		"max_tokens": 1,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(respBody))
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
		var citations []Citation

		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and non-data lines
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// Check for stream end
			if data == "[DONE]" {
				// If we have citations, append them to the final chunk
				if len(citations) > 0 {
					chunks <- llm.StreamChunk{
						Delta:        c.formatCitations(citations),
						FinishReason: "stop",
					}
				}
				break
			}

			// Parse the SSE data
			var streamResp streamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue
			}

			// Capture citations from the response
			if len(streamResp.Citations) > 0 {
				citations = streamResp.Citations
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

			// Check for finish reason
			if choice.FinishReason != "" && choice.FinishReason != "null" {
				// Don't emit finish reason yet if we have citations to append
				if len(citations) == 0 {
					chunks <- llm.StreamChunk{
						FinishReason: choice.FinishReason,
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

// formatCitations formats citations as markdown
func (c *Client) formatCitations(citations []Citation) string {
	if len(citations) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\n---\n**Sources:**\n")
	for i, citation := range citations {
		sb.WriteString(fmt.Sprintf("%d. [%s](%s)\n", i+1, citation.Title, citation.URL))
	}
	return sb.String()
}

// convertMessages converts llm.Message to Perplexity/OpenAI format
func (c *Client) convertMessages(messages []llm.Message) []map[string]interface{} {
	result := make([]map[string]interface{}, len(messages))

	for i, msg := range messages {
		m := map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
		result[i] = m
	}

	return result
}

// Citation represents a citation/source from Perplexity's response
type Citation struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// streamResponse represents a Perplexity streaming response
type streamResponse struct {
	ID        string     `json:"id"`
	Object    string     `json:"object"`
	Created   int64      `json:"created"`
	Model     string     `json:"model"`
	Citations []Citation `json:"citations,omitempty"`
	Choices   []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}
