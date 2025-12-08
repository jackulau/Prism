package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jacklau/prism/internal/integrations"
)

// Config holds Discord integration configuration
type Config struct {
	WebhookURL string
	BotToken   string
	Enabled    bool
}

// Client is a Discord notification client
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new Discord client
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
	return "discord"
}

// Enabled returns whether the provider is enabled
func (c *Client) Enabled() bool {
	return c.config.Enabled && c.config.WebhookURL != ""
}

// Send sends a notification to Discord
func (c *Client) Send(event *integrations.Event) error {
	if !c.Enabled() {
		return nil
	}

	// Build the message based on event type
	message := c.buildMessage(event)

	// Create webhook payload
	payload := map[string]interface{}{
		"content": message,
		"embeds":  c.buildEmbeds(event),
	}

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
		return fmt.Errorf("discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// buildMessage builds the notification message
func (c *Client) buildMessage(event *integrations.Event) string {
	switch event.Type {
	case integrations.EventError:
		return "⚠️ **Error Alert**"
	case integrations.EventToolApproved:
		return "✅ **Tool Approved**"
	case integrations.EventToolRejected:
		return "❌ **Tool Rejected**"
	default:
		return fmt.Sprintf("📋 **Event: %s**", event.Type)
	}
}

// buildEmbeds builds Discord embeds for the event
func (c *Client) buildEmbeds(event *integrations.Event) []map[string]interface{} {
	embed := map[string]interface{}{
		"title":     string(event.Type),
		"timestamp": time.Now().Format(time.RFC3339),
		"fields":    []map[string]interface{}{},
	}

	// Set color based on event type
	switch event.Type {
	case integrations.EventError:
		embed["color"] = 15158332 // Red
	case integrations.EventToolApproved:
		embed["color"] = 3066993 // Green
	case integrations.EventToolRejected:
		embed["color"] = 15105570 // Orange
	default:
		embed["color"] = 3447003 // Blue
	}

	// Add fields
	fields := []map[string]interface{}{}

	if event.UserID != "" {
		fields = append(fields, map[string]interface{}{
			"name":   "User ID",
			"value":  event.UserID,
			"inline": true,
		})
	}

	if event.ConversationID != "" {
		fields = append(fields, map[string]interface{}{
			"name":   "Conversation ID",
			"value":  event.ConversationID,
			"inline": true,
		})
	}

	// Add custom data fields
	if event.Data != nil {
		for key, value := range event.Data {
			fields = append(fields, map[string]interface{}{
				"name":   key,
				"value":  fmt.Sprintf("%v", value),
				"inline": true,
			})
		}
	}

	embed["fields"] = fields

	return []map[string]interface{}{embed}
}
