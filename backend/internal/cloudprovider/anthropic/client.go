package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/jacklau/prism/internal/cloudprovider"
)

const (
	defaultBaseURL   = "https://api.anthropic.com/v1"
	anthropicVersion = "2023-06-01"
	defaultMaxTokens = 4096
	defaultModel     = "claude-sonnet-4-5-20250929"
)

// Client implements the CloudProvider interface for Anthropic
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client

	// In-memory agent storage (in production, this would be a database)
	agents   map[string]*cloudprovider.Agent
	agentsMu sync.RWMutex

	// Message history per agent
	messages   map[string][]cloudprovider.ProviderMessage
	messagesMu sync.RWMutex
}

// NewClient creates a new Anthropic cloud provider client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		agents:     make(map[string]*cloudprovider.Agent),
		messages:   make(map[string][]cloudprovider.ProviderMessage),
	}
}

// NewClientWithConfig creates a new client with custom configuration
func NewClientWithConfig(config cloudprovider.ProviderConfig) *Client {
	baseURL := defaultBaseURL
	if config.BaseURL != "" {
		baseURL = config.BaseURL
	}

	return &Client{
		apiKey:     config.APIKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		agents:     make(map[string]*cloudprovider.Agent),
		messages:   make(map[string][]cloudprovider.ProviderMessage),
	}
}

// Name returns the provider name
func (c *Client) Name() string {
	return "anthropic-cloud"
}

// HasCredentials returns whether credentials are configured
func (c *Client) HasCredentials() bool {
	return c.apiKey != ""
}

// ValidateCredentials validates the provider credentials
func (c *Client) ValidateCredentials(ctx context.Context) error {
	if c.apiKey == "" {
		return cloudprovider.ErrNoCredentials
	}

	// Make a minimal request to validate the API key
	body := map[string]interface{}{
		"model":      "claude-haiku-4-5-20251001",
		"max_tokens": 1,
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hi"},
		},
	}

	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate credentials: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return cloudprovider.ErrUnauthorized
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return cloudprovider.ErrRateLimited
	}

	if resp.StatusCode >= 500 {
		return cloudprovider.ErrProviderUnavailable
	}

	return nil
}

// CreateAgent creates a new agent with the given parameters
func (c *Client) CreateAgent(ctx context.Context, params cloudprovider.CreateAgentParams) (*cloudprovider.Agent, error) {
	if c.apiKey == "" {
		return nil, cloudprovider.ErrNoCredentials
	}

	// Generate a unique agent ID
	agentID := fmt.Sprintf("agent_%d", time.Now().UnixNano())

	model := params.Model
	if model == "" {
		model = defaultModel
	}

	agent := &cloudprovider.Agent{
		ID:           agentID,
		ProviderID:   agentID,
		Name:         params.Name,
		Model:        model,
		Status:       cloudprovider.AgentStatusActive,
		SystemPrompt: params.SystemPrompt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Metadata:     params.Metadata,
	}

	c.agentsMu.Lock()
	c.agents[agentID] = agent
	c.agentsMu.Unlock()

	c.messagesMu.Lock()
	c.messages[agentID] = []cloudprovider.ProviderMessage{}
	c.messagesMu.Unlock()

	return agent, nil
}

// GetAgent retrieves an agent by ID
func (c *Client) GetAgent(ctx context.Context, agentID string) (*cloudprovider.Agent, error) {
	c.agentsMu.RLock()
	agent, exists := c.agents[agentID]
	c.agentsMu.RUnlock()

	if !exists {
		return nil, cloudprovider.ErrAgentNotFound
	}

	return agent, nil
}

// DeleteAgent removes an agent
func (c *Client) DeleteAgent(ctx context.Context, agentID string) error {
	c.agentsMu.Lock()
	_, exists := c.agents[agentID]
	if exists {
		delete(c.agents, agentID)
	}
	c.agentsMu.Unlock()

	if !exists {
		return cloudprovider.ErrAgentNotFound
	}

	c.messagesMu.Lock()
	delete(c.messages, agentID)
	c.messagesMu.Unlock()

	return nil
}

// GetMessages retrieves all messages for an agent's conversation
func (c *Client) GetMessages(ctx context.Context, agent *cloudprovider.Agent) ([]cloudprovider.ProviderMessage, error) {
	c.messagesMu.RLock()
	msgs, exists := c.messages[agent.ID]
	c.messagesMu.RUnlock()

	if !exists {
		return nil, cloudprovider.ErrAgentNotFound
	}

	// Return a copy to prevent mutation
	result := make([]cloudprovider.ProviderMessage, len(msgs))
	copy(result, msgs)
	return result, nil
}

// SendMessage sends a message to an agent and returns success status
func (c *Client) SendMessage(ctx context.Context, agent *cloudprovider.Agent, message string, images []cloudprovider.ImageData) (bool, error) {
	if c.apiKey == "" {
		return false, cloudprovider.ErrNoCredentials
	}

	// Build the message request
	req := c.buildMessageRequest(agent, message, images, false)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return false, err
	}

	// Parse response
	var msgResp messageResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Store messages
	c.storeUserMessage(agent.ID, message, images)
	c.storeAssistantMessage(agent.ID, &msgResp)

	return true, nil
}

// StreamMessages returns a channel for streaming agent responses
func (c *Client) StreamMessages(ctx context.Context, agent *cloudprovider.Agent) (<-chan cloudprovider.MessageChunk, error) {
	if c.apiKey == "" {
		return nil, cloudprovider.ErrNoCredentials
	}

	// Get the last user message to stream a response
	c.messagesMu.RLock()
	msgs := c.messages[agent.ID]
	c.messagesMu.RUnlock()

	if len(msgs) == 0 {
		return nil, cloudprovider.ErrInvalidRequest
	}

	// Build a streaming request with the conversation history
	req := c.buildStreamRequest(agent)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if err := c.checkResponse(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}

	// Create streaming channel
	chunks := make(chan cloudprovider.MessageChunk, 100)

	// Start streaming in background
	reader := newStreamReader(ctx, resp.Body)
	go reader.processStream(chunks)

	return chunks, nil
}

// Helper methods

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) checkResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)

	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
		return cloudprovider.NewAPIError(resp.StatusCode, errResp.Error.Type, errResp.Error.Message)
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return cloudprovider.ErrUnauthorized
	case http.StatusTooManyRequests:
		return cloudprovider.ErrRateLimited
	case http.StatusNotFound:
		return cloudprovider.ErrAgentNotFound
	default:
		return cloudprovider.NewAPIError(resp.StatusCode, "", string(body))
	}
}

func (c *Client) buildMessageRequest(agent *cloudprovider.Agent, message string, images []cloudprovider.ImageData, stream bool) *sendMessageRequest {
	// Build content parts
	var content []contentPart

	// Add images first
	content = append(content, fromImages(images)...)

	// Add text message
	content = append(content, contentPart{
		Type: "text",
		Text: message,
	})

	// Build messages from history
	var messages []messageRequest

	c.messagesMu.RLock()
	history := c.messages[agent.ID]
	c.messagesMu.RUnlock()

	for _, msg := range history {
		var msgContent []contentPart
		if msg.Content != "" {
			msgContent = append(msgContent, contentPart{
				Type: "text",
				Text: msg.Content,
			})
		}
		for _, tc := range msg.ToolCalls {
			msgContent = append(msgContent, contentPart{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Name,
				Input: tc.Parameters,
			})
		}
		if msg.ToolCallID != "" {
			msgContent = []contentPart{{
				Type:      "tool_result",
				ToolUseID: msg.ToolCallID,
				Content:   msg.Content,
			}}
		}
		messages = append(messages, messageRequest{
			Role:    msg.Role,
			Content: msgContent,
		})
	}

	// Add new user message
	messages = append(messages, messageRequest{
		Role:    "user",
		Content: content,
	})

	return &sendMessageRequest{
		Model:     agent.Model,
		MaxTokens: defaultMaxTokens,
		Messages:  messages,
		System:    agent.SystemPrompt,
		Stream:    stream,
	}
}

func (c *Client) buildStreamRequest(agent *cloudprovider.Agent) *sendMessageRequest {
	// Build messages from history
	var messages []messageRequest

	c.messagesMu.RLock()
	history := c.messages[agent.ID]
	c.messagesMu.RUnlock()

	for _, msg := range history {
		var msgContent []contentPart
		if msg.Content != "" {
			msgContent = append(msgContent, contentPart{
				Type: "text",
				Text: msg.Content,
			})
		}
		for _, tc := range msg.ToolCalls {
			msgContent = append(msgContent, contentPart{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Name,
				Input: tc.Parameters,
			})
		}
		if msg.ToolCallID != "" {
			msgContent = []contentPart{{
				Type:      "tool_result",
				ToolUseID: msg.ToolCallID,
				Content:   msg.Content,
			}}
		}
		messages = append(messages, messageRequest{
			Role:    msg.Role,
			Content: msgContent,
		})
	}

	return &sendMessageRequest{
		Model:     agent.Model,
		MaxTokens: defaultMaxTokens,
		Messages:  messages,
		System:    agent.SystemPrompt,
		Stream:    true,
	}
}

func (c *Client) storeUserMessage(agentID, message string, images []cloudprovider.ImageData) {
	msg := cloudprovider.ProviderMessage{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	}

	c.messagesMu.Lock()
	c.messages[agentID] = append(c.messages[agentID], msg)
	c.messagesMu.Unlock()
}

func (c *Client) storeAssistantMessage(agentID string, resp *messageResponse) {
	msgs := resp.toProviderMessages()
	if len(msgs) > 0 {
		c.messagesMu.Lock()
		c.messages[agentID] = append(c.messages[agentID], msgs...)
		c.messagesMu.Unlock()
	}
}

// Capabilities returns the capabilities of this provider
func (c *Client) Capabilities() cloudprovider.ProviderCapabilities {
	return cloudprovider.ProviderCapabilities{
		SupportsTools:     true,
		SupportsVision:    true,
		SupportsStreaming: true,
		SupportedModels: []string{
			"claude-sonnet-4-5-20250929",
			"claude-haiku-4-5-20251001",
			"claude-opus-4-5-20250929",
		},
		MaxContextWindow: 200000,
	}
}

// Ensure Client implements CloudProvider
var _ cloudprovider.CloudProvider = (*Client)(nil)
