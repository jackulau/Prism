package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jacklau/prism/internal/llm"
)

// ImageGenerationTool generates images using OpenAI's DALL-E API
type ImageGenerationTool struct {
	openaiAPIKey string
	httpClient   *http.Client
}

// NewImageGenerationTool creates a new image generation tool
func NewImageGenerationTool(openaiAPIKey string) *ImageGenerationTool {
	return &ImageGenerationTool{
		openaiAPIKey: openaiAPIKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (t *ImageGenerationTool) Name() string {
	return "generate_image"
}

func (t *ImageGenerationTool) Description() string {
	return "Generate an image based on a text description using DALL-E. Returns a URL to the generated image. Use descriptive prompts for best results."
}

func (t *ImageGenerationTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"prompt": {
				Type:        "string",
				Description: "A detailed description of the image to generate",
			},
			"size": {
				Type:        "string",
				Description: "The size of the generated image",
				Enum:        []string{"256x256", "512x512", "1024x1024", "1792x1024", "1024x1792"},
				Default:     "1024x1024",
			},
			"quality": {
				Type:        "string",
				Description: "The quality of the generated image",
				Enum:        []string{"standard", "hd"},
				Default:     "standard",
			},
			"style": {
				Type:        "string",
				Description: "The style of the generated image",
				Enum:        []string{"vivid", "natural"},
				Default:     "vivid",
			},
		},
		Required: []string{"prompt"},
	}
}

func (t *ImageGenerationTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	if t.openaiAPIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	prompt, ok := params["prompt"].(string)
	if !ok || prompt == "" {
		return nil, fmt.Errorf("prompt parameter is required")
	}

	size := "1024x1024"
	if s, ok := params["size"].(string); ok && s != "" {
		size = s
	}

	quality := "standard"
	if q, ok := params["quality"].(string); ok && q != "" {
		quality = q
	}

	style := "vivid"
	if s, ok := params["style"].(string); ok && s != "" {
		style = s
	}

	// Create request body
	reqBody := map[string]interface{}{
		"model":   "dall-e-3",
		"prompt":  prompt,
		"n":       1,
		"size":    size,
		"quality": quality,
		"style":   style,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/images/generations", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.openaiAPIKey)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var openaiResp struct {
		Created int64 `json:"created"`
		Data    []struct {
			URL           string `json:"url"`
			RevisedPrompt string `json:"revised_prompt"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(openaiResp.Data) == 0 {
		return nil, fmt.Errorf("no images generated")
	}

	return map[string]interface{}{
		"url":            openaiResp.Data[0].URL,
		"revised_prompt": openaiResp.Data[0].RevisedPrompt,
		"size":           size,
	}, nil
}

func (t *ImageGenerationTool) RequiresConfirmation() bool {
	return true // Image generation costs money and should require confirmation
}
