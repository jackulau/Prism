package cursor

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jacklau/prism/internal/providers"
)

// HTTPClient handles HTTP communication with Cursor API
type HTTPClient struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

// NewHTTPClient creates a new HTTP client for Cursor API
func NewHTTPClient(apiKey string, baseURL string) *HTTPClient {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

// SetAPIKey updates the API key
func (c *HTTPClient) SetAPIKey(key string) {
	c.apiKey = key
}

// HasAPIKey returns whether an API key is configured
func (c *HTTPClient) HasAPIKey() bool {
	return c.apiKey != ""
}

// buildAuthHeader creates the Basic auth header value
// Cursor uses Basic auth with the API key Base64 encoded: base64(api_key + ":")
func (c *HTTPClient) buildAuthHeader() string {
	credentials := c.apiKey + ":"
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return "Basic " + encoded
}

// doRequest performs an HTTP request with authentication
func (c *HTTPClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.buildAuthHeader())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return c.client.Do(req)
}

// doStreamingRequest performs an HTTP request expecting SSE response
func (c *HTTPClient) doStreamingRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.buildAuthHeader())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	return c.client.Do(req)
}

// parseErrorResponse parses an error response from Cursor API
func (c *HTTPClient) parseErrorResponse(resp *http.Response) *providers.ProviderError {
	body, _ := io.ReadAll(resp.Body)

	var errResp CursorErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
		return &providers.ProviderError{
			Provider:   "cursor",
			Code:       errResp.Error.Code,
			Message:    errResp.Error.Message,
			StatusCode: resp.StatusCode,
			Retryable:  isRetryableStatus(resp.StatusCode),
		}
	}

	// Fallback for non-JSON error responses
	return &providers.ProviderError{
		Provider:   "cursor",
		Code:       fmt.Sprintf("http_%d", resp.StatusCode),
		Message:    fmt.Sprintf("API error: %s", string(body)),
		StatusCode: resp.StatusCode,
		Retryable:  isRetryableStatus(resp.StatusCode),
	}
}

// isRetryableStatus returns whether an HTTP status code indicates a retryable error
func isRetryableStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests, // 429
		http.StatusBadGateway,        // 502
		http.StatusServiceUnavailable, // 503
		http.StatusGatewayTimeout:     // 504
		return true
	default:
		return false
	}
}

// CreateAgent creates a new agent via Cursor API
func (c *HTTPClient) CreateAgent(ctx context.Context, req CursorCreateAgentRequest) (*CursorAgentResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/agents", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseErrorResponse(resp)
	}

	var agentResp CursorAgentResponse
	if err := json.NewDecoder(resp.Body).Decode(&agentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &agentResp, nil
}

// GetAgent retrieves an agent by ID
func (c *HTTPClient) GetAgent(ctx context.Context, agentID string) (*CursorAgentResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/agents/"+agentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseErrorResponse(resp)
	}

	var agentResp CursorAgentResponse
	if err := json.NewDecoder(resp.Body).Decode(&agentResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &agentResp, nil
}

// GetMessages retrieves messages for an agent
func (c *HTTPClient) GetMessages(ctx context.Context, agentID string) (*CursorMessagesResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/agents/"+agentID+"/messages", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseErrorResponse(resp)
	}

	var messagesResp CursorMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&messagesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &messagesResp, nil
}

// SendFollowup sends a follow-up message and returns the response for streaming
func (c *HTTPClient) SendFollowup(ctx context.Context, agentID, message string) (*http.Response, error) {
	req := CursorFollowupRequest{
		Message: message,
	}

	resp, err := c.doStreamingRequest(ctx, http.MethodPost, "/agents/"+agentID+"/followup", req)
	if err != nil {
		return nil, fmt.Errorf("failed to send followup: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, c.parseErrorResponse(resp)
	}

	return resp, nil
}

// StopAgent stops a running agent
func (c *HTTPClient) StopAgent(ctx context.Context, agentID string) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/agents/"+agentID+"/stop", nil)
	if err != nil {
		return fmt.Errorf("failed to stop agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseErrorResponse(resp)
	}

	return nil
}

// ValidateKey validates an API key by making a test request
func (c *HTTPClient) ValidateKey(ctx context.Context, key string) error {
	// Temporarily set the key for validation
	originalKey := c.apiKey
	c.apiKey = key
	defer func() { c.apiKey = originalKey }()

	// Try a lightweight request to validate the key
	// Most APIs return 401 for invalid keys on any authenticated endpoint
	resp, err := c.doRequest(ctx, http.MethodGet, "/agents", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}

	// Any non-error response means the key is valid
	return nil
}
