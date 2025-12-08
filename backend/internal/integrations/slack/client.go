package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jacklau/prism/internal/integrations"
)

// Config holds Slack integration configuration
type Config struct {
	WebhookURL string
	BotToken   string
	ChannelID  string
	Enabled    bool
}

// Client is a Slack notification client
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new Slack client
func NewClient(config *Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name returns the provider name
func (c *Client) Name() string {
	return "slack"
}

// Enabled returns whether the provider is enabled
func (c *Client) Enabled() bool {
	return c.config.Enabled && c.config.WebhookURL != ""
}

// Send sends a notification to Slack
func (c *Client) Send(event *integrations.Event) error {
	if !c.Enabled() {
		return nil
	}

	// Build the message payload
	payload := c.buildPayload(event)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.WebhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// buildPayload builds the Slack message payload
func (c *Client) buildPayload(event *integrations.Event) map[string]interface{} {
	// Build blocks for rich formatting
	blocks := []map[string]interface{}{}

	// Header block
	headerText := c.getHeaderText(event)
	blocks = append(blocks, map[string]interface{}{
		"type": "header",
		"text": map[string]interface{}{
			"type":  "plain_text",
			"text":  headerText,
			"emoji": true,
		},
	})

	// Context block with event details
	contextElements := []map[string]interface{}{}

	if event.UserID != "" {
		contextElements = append(contextElements, map[string]interface{}{
			"type": "mrkdwn",
			"text": fmt.Sprintf("*User:* %s", event.UserID),
		})
	}

	if event.ConversationID != "" {
		contextElements = append(contextElements, map[string]interface{}{
			"type": "mrkdwn",
			"text": fmt.Sprintf("*Conversation:* %s", event.ConversationID),
		})
	}

	if len(contextElements) > 0 {
		blocks = append(blocks, map[string]interface{}{
			"type":     "context",
			"elements": contextElements,
		})
	}

	// Section block for additional data
	if event.Data != nil && len(event.Data) > 0 {
		fields := []map[string]interface{}{}
		for key, value := range event.Data {
			fields = append(fields, map[string]interface{}{
				"type": "mrkdwn",
				"text": fmt.Sprintf("*%s:*\n%v", key, value),
			})
		}

		blocks = append(blocks, map[string]interface{}{
			"type":   "section",
			"fields": fields,
		})
	}

	// Divider
	blocks = append(blocks, map[string]interface{}{
		"type": "divider",
	})

	return map[string]interface{}{
		"blocks": blocks,
		"text":   headerText, // Fallback text
	}
}

// getHeaderText returns the header text based on event type
func (c *Client) getHeaderText(event *integrations.Event) string {
	switch event.Type {
	case integrations.EventError:
		return "⚠️ Error Alert"
	case integrations.EventToolApproved:
		return "✅ Tool Approved"
	case integrations.EventToolRejected:
		return "❌ Tool Rejected"
	case integrations.EventChatCompleted:
		return "💬 Chat Completed"
	case integrations.EventChatStopped:
		return "🛑 Chat Stopped"
	case integrations.EventToolStarted:
		return "🔧 Tool Started"
	case integrations.EventToolCompleted:
		return "✔️ Tool Completed"
	default:
		return fmt.Sprintf("📋 Event: %s", event.Type)
	}
}
