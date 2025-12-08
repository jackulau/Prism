package posthog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jacklau/prism/internal/integrations"
)

// Config holds PostHog integration configuration
type Config struct {
	APIKey        string
	Endpoint      string
	Enabled       bool
	BatchSize     int
	FlushInterval time.Duration
}

// Client is a PostHog analytics client
type Client struct {
	config     *Config
	httpClient *http.Client
	queue      []event
	mu         sync.Mutex
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

// event represents a PostHog event
type event struct {
	Event      string                 `json:"event"`
	DistinctID string                 `json:"distinct_id"`
	Properties map[string]interface{} `json:"properties"`
	Timestamp  time.Time              `json:"timestamp"`
}

// NewClient creates a new PostHog client
func NewClient(config *Config) *Client {
	if config.Endpoint == "" {
		config.Endpoint = "https://app.posthog.com"
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 30 * time.Second
	}

	c := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		queue:  make([]event, 0, config.BatchSize),
		stopCh: make(chan struct{}),
	}

	// Start background flusher if enabled
	if config.Enabled && config.APIKey != "" {
		c.wg.Add(1)
		go c.backgroundFlusher()
	}

	return c
}

// Name returns the provider name
func (c *Client) Name() string {
	return "posthog"
}

// Enabled returns whether the provider is enabled
func (c *Client) Enabled() bool {
	return c.config.Enabled && c.config.APIKey != ""
}

// Track tracks an event
func (c *Client) Track(evt *integrations.Event) error {
	if !c.Enabled() {
		return nil
	}

	properties := map[string]interface{}{
		"event_type":      string(evt.Type),
		"conversation_id": evt.ConversationID,
		"message_id":      evt.MessageID,
	}

	// Merge custom data
	for key, value := range evt.Data {
		properties[key] = value
	}

	c.mu.Lock()
	c.queue = append(c.queue, event{
		Event:      string(evt.Type),
		DistinctID: evt.UserID,
		Properties: properties,
		Timestamp:  time.Now(),
	})

	// Flush if queue is full
	if len(c.queue) >= c.config.BatchSize {
		c.mu.Unlock()
		return c.Flush()
	}
	c.mu.Unlock()

	return nil
}

// Flush flushes the event queue
func (c *Client) Flush() error {
	if !c.Enabled() {
		return nil
	}

	c.mu.Lock()
	if len(c.queue) == 0 {
		c.mu.Unlock()
		return nil
	}

	events := make([]event, len(c.queue))
	copy(events, c.queue)
	c.queue = c.queue[:0]
	c.mu.Unlock()

	return c.sendBatch(events)
}

// Close closes the client
func (c *Client) Close() error {
	if c.config.Enabled && c.config.APIKey != "" {
		close(c.stopCh)
		c.wg.Wait()
	}

	// Final flush
	return c.Flush()
}

// backgroundFlusher periodically flushes the event queue
func (c *Client) backgroundFlusher() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.Flush(); err != nil {
				// Log error but continue
				fmt.Printf("PostHog flush error: %v\n", err)
			}
		case <-c.stopCh:
			return
		}
	}
}

// sendBatch sends a batch of events to PostHog
func (c *Client) sendBatch(events []event) error {
	payload := map[string]interface{}{
		"api_key": c.config.APIKey,
		"batch":   events,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/batch/", c.config.Endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
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
		return fmt.Errorf("posthog returned status %d", resp.StatusCode)
	}

	return nil
}

// Identify identifies a user with properties
func (c *Client) Identify(userID string, properties map[string]interface{}) error {
	if !c.Enabled() {
		return nil
	}

	payload := map[string]interface{}{
		"api_key":     c.config.APIKey,
		"distinct_id": userID,
		"properties":  properties,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/identify/", c.config.Endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
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
		return fmt.Errorf("posthog returned status %d", resp.StatusCode)
	}

	return nil
}
