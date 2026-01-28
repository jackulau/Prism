package openai

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jacklau/prism/internal/llm"
)

// ImageDetail specifies the detail level for image processing
type ImageDetail string

const (
	ImageDetailLow  ImageDetail = "low"
	ImageDetailHigh ImageDetail = "high"
	ImageDetailAuto ImageDetail = "auto"
)

// ImageContent represents image content in OpenAI format
type ImageContent struct {
	Type     string    `json:"type"` // "image_url"
	ImageURL *ImageURL `json:"image_url"`
}

// ImageURL represents an image URL with detail level
type ImageURL struct {
	URL    string `json:"url"`    // base64 data URL or HTTP URL
	Detail string `json:"detail"` // "low", "high", "auto"
}

// TextContent represents text content in OpenAI format
type TextContent struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
}

// formatVisionMessage converts a message with images to OpenAI format
func (c *Client) formatVisionMessage(msg *llm.Message) (map[string]interface{}, error) {
	if len(msg.Images) == 0 {
		// No images, return simple message format
		return map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}, nil
	}

	// Build multimodal content array
	content := make([]interface{}, 0, len(msg.Images)+1)

	// Add text content if present
	if msg.Content != "" {
		content = append(content, TextContent{
			Type: "text",
			Text: msg.Content,
		})
	}

	// Add image content
	for _, img := range msg.Images {
		imageURL, err := formatImageURL(img)
		if err != nil {
			return nil, err
		}
		content = append(content, ImageContent{
			Type:     "image_url",
			ImageURL: imageURL,
		})
	}

	return map[string]interface{}{
		"role":    msg.Role,
		"content": content,
	}, nil
}

// formatImageURL formats an ImageData to an ImageURL
func formatImageURL(img llm.ImageData) (*ImageURL, error) {
	var url string

	if img.URL != "" {
		// Use URL directly
		url = img.URL
	} else if img.Base64 != "" {
		// Build data URL from base64
		mimeType := img.MimeType
		if mimeType == "" {
			// Try to detect from base64 data
			mimeType = detectImageMimeType(img.Base64)
		}
		if mimeType == "" {
			mimeType = "image/png" // Default to PNG
		}
		url = fmt.Sprintf("data:%s;base64,%s", mimeType, img.Base64)
	} else {
		return nil, fmt.Errorf("image must have either URL or Base64 data")
	}

	return &ImageURL{
		URL:    url,
		Detail: "auto", // Default to auto
	}, nil
}

// detectImageMimeType attempts to detect the MIME type from base64 data
func detectImageMimeType(base64Data string) string {
	// Decode first few bytes to check magic numbers
	length := len(base64Data)
	if length > 100 {
		length = 100
	}
	data, err := base64.StdEncoding.DecodeString(base64Data[:length])
	if err != nil {
		return ""
	}

	// Check magic numbers
	if len(data) >= 8 {
		// PNG: 89 50 4E 47 0D 0A 1A 0A
		if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
			return "image/png"
		}
		// JPEG: FF D8 FF
		if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
			return "image/jpeg"
		}
		// GIF: 47 49 46 38
		if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 {
			return "image/gif"
		}
		// WebP: 52 49 46 46 ... 57 45 42 50
		if data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
			if len(data) >= 12 && data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
				return "image/webp"
			}
		}
	}

	return ""
}

// convertMessagesWithVision converts messages to OpenAI format, handling images
func (c *Client) convertMessagesWithVision(messages []llm.Message) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(messages))

	for _, msg := range messages {
		converted, err := c.formatVisionMessage(&msg)
		if err != nil {
			return nil, fmt.Errorf("failed to format message: %w", err)
		}

		// Handle tool calls
		if len(msg.ToolCalls) > 0 {
			toolCalls := make([]map[string]interface{}, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				toolCalls[j] = map[string]interface{}{
					"id":   tc.ID,
					"type": "function",
					"function": map[string]interface{}{
						"name":      tc.Name,
						"arguments": mustMarshalJSON(tc.Parameters),
					},
				}
			}
			converted["tool_calls"] = toolCalls
		}

		if msg.ToolCallID != "" {
			converted["tool_call_id"] = msg.ToolCallID
		}

		result = append(result, converted)
	}

	return result, nil
}

// hasImages checks if any message contains images
func hasImages(messages []llm.Message) bool {
	for _, msg := range messages {
		if len(msg.Images) > 0 {
			return true
		}
	}
	return false
}

// ValidateVisionSupport checks if the model supports vision
func ValidateVisionSupport(modelID string) error {
	config := GetModelConfig(modelID)
	if !config.SupportsVision {
		return fmt.Errorf("model %s does not support vision", modelID)
	}
	return nil
}

// mustMarshalJSON marshals to JSON string, panics on error (for internal use only)
func mustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// IsValidImageURL checks if a string is a valid image URL
func IsValidImageURL(url string) bool {
	// Check for data URLs
	if strings.HasPrefix(url, "data:image/") {
		return true
	}
	// Check for HTTP URLs
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		// Simple extension check
		lower := strings.ToLower(url)
		return strings.Contains(lower, ".png") ||
			strings.Contains(lower, ".jpg") ||
			strings.Contains(lower, ".jpeg") ||
			strings.Contains(lower, ".gif") ||
			strings.Contains(lower, ".webp")
	}
	return false
}
