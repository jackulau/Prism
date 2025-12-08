package google

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

// Client implements the LLM provider interface for Google AI (Gemini)
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewClient creates a new Google AI client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
		client:  &http.Client{},
	}
}

// Name returns the provider name
func (c *Client) Name() string {
	return "google"
}

// Models returns available models
func (c *Client) Models() []llm.Model {
	return []llm.Model{
		{
			ID:             "gemini-2.0-flash-exp",
			Name:           "Gemini 2.0 Flash",
			Description:    "Latest multimodal model with enhanced capabilities",
			ContextWindow:  1000000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "gemini-1.5-pro",
			Name:           "Gemini 1.5 Pro",
			Description:    "Best for complex reasoning tasks",
			ContextWindow:  2000000,
			SupportsTools:  true,
			SupportsVision: true,
		},
		{
			ID:             "gemini-1.5-flash",
			Name:           "Gemini 1.5 Flash",
			Description:    "Fast and versatile for most tasks",
			ContextWindow:  1000000,
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
	url := fmt.Sprintf("%s/models?key=%s", c.baseURL, key)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

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

// Chat sends a chat request and returns a streaming response
func (c *Client) Chat(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	// Convert messages to Gemini format
	contents := c.convertMessages(req.Messages)

	// Build request body
	body := map[string]interface{}{
		"contents": contents,
	}

	// Add generation config
	generationConfig := map[string]interface{}{}
	if req.MaxTokens > 0 {
		generationConfig["maxOutputTokens"] = req.MaxTokens
	}
	if req.Temperature > 0 {
		generationConfig["temperature"] = req.Temperature
	}
	if len(generationConfig) > 0 {
		body["generationConfig"] = generationConfig
	}

	// Add tools
	if len(req.Tools) > 0 {
		body["tools"] = []map[string]interface{}{
			{
				"functionDeclarations": c.convertTools(req.Tools),
			},
		}
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s&alt=sse", c.baseURL, req.Model, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
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
			line := scanner.Text()

			// Skip empty lines
			if line == "" {
				continue
			}

			// Parse SSE data
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			var streamResp geminiStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue
			}

			for _, candidate := range streamResp.Candidates {
				for _, part := range candidate.Content.Parts {
					if part.Text != "" {
						chunks <- llm.StreamChunk{
							Delta: part.Text,
						}
					}

					if part.FunctionCall != nil {
						chunks <- llm.StreamChunk{
							ToolCalls: []llm.ToolCall{
								{
									ID:         part.FunctionCall.Name, // Gemini doesn't provide IDs
									Name:       part.FunctionCall.Name,
									Parameters: part.FunctionCall.Args,
								},
							},
						}
					}
				}

				if candidate.FinishReason != "" {
					finishReason := strings.ToLower(candidate.FinishReason)
					if finishReason == "stop" || finishReason == "end_turn" {
						finishReason = "stop"
					}
					chunks <- llm.StreamChunk{
						FinishReason: finishReason,
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

// convertMessages converts llm.Message to Gemini format
func (c *Client) convertMessages(messages []llm.Message) []map[string]interface{} {
	var contents []map[string]interface{}

	for _, msg := range messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		if role == "system" {
			// Gemini handles system messages differently
			// We'll prepend it to the first user message
			continue
		}

		parts := []map[string]interface{}{}

		// Handle images
		for _, img := range msg.Images {
			if img.Base64 != "" {
				parts = append(parts, map[string]interface{}{
					"inlineData": map[string]interface{}{
						"mimeType": img.MimeType,
						"data":     img.Base64,
					},
				})
			}
		}

		// Handle text
		if msg.Content != "" {
			parts = append(parts, map[string]interface{}{
				"text": msg.Content,
			})
		}

		// Handle tool responses
		if msg.ToolCallID != "" {
			parts = []map[string]interface{}{
				{
					"functionResponse": map[string]interface{}{
						"name": msg.ToolCallID,
						"response": map[string]interface{}{
							"content": msg.Content,
						},
					},
				},
			}
		}

		contents = append(contents, map[string]interface{}{
			"role":  role,
			"parts": parts,
		})
	}

	return contents
}

// convertTools converts llm.ToolDefinition to Gemini format
func (c *Client) convertTools(tools []llm.ToolDefinition) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tools))

	for i, tool := range tools {
		result[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"parameters":  tool.Parameters,
		}
	}

	return result
}

// geminiStreamResponse represents a Gemini streaming response
type geminiStreamResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text         string `json:"text"`
				FunctionCall *struct {
					Name string                 `json:"name"`
					Args map[string]interface{} `json:"args"`
				} `json:"functionCall"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason  string `json:"finishReason"`
		SafetyRatings []struct {
			Category    string `json:"category"`
			Probability string `json:"probability"`
		} `json:"safetyRatings"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}
