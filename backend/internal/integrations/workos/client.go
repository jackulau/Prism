package workos

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultBaseURL = "https://api.workos.com"
	apiVersion     = "2023-01-01"
)

// Config holds WorkOS client configuration
type Config struct {
	APIKey        string
	ClientID      string
	WebhookSecret string
	BaseURL       string
	Enabled       bool
}

// Client is a WorkOS API client
type Client struct {
	config     *Config
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new WorkOS client
func NewClient(config *Config) *Client {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		config:  config,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Enabled returns whether the client is enabled and configured
func (c *Client) Enabled() bool {
	return c.config.Enabled && c.config.APIKey != ""
}

// CreateOrganization creates a new organization in WorkOS
func (c *Client) CreateOrganization(name string) (*Organization, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("workos client is not enabled")
	}

	req := CreateOrganizationRequest{
		Name: name,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/organizations", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, c.parseError(resp)
	}

	var org Organization
	if err := json.NewDecoder(resp.Body).Decode(&org); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("WorkOS: Created organization %s (%s)", org.Name, org.ID)
	return &org, nil
}

// GetOrganization retrieves an organization from WorkOS by ID
func (c *Client) GetOrganization(workosOrgID string) (*Organization, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("workos client is not enabled")
	}

	httpReq, err := http.NewRequest("GET", c.baseURL+"/organizations/"+workosOrgID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, nil
	}

	if resp.StatusCode >= 400 {
		return nil, c.parseError(resp)
	}

	var org Organization
	if err := json.NewDecoder(resp.Body).Decode(&org); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &org, nil
}

// UpdateOrganization updates an organization in WorkOS
func (c *Client) UpdateOrganization(workosOrgID, name string) error {
	if !c.Enabled() {
		return fmt.Errorf("workos client is not enabled")
	}

	req := UpdateOrganizationRequest{
		Name: name,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("PUT", c.baseURL+"/organizations/"+workosOrgID, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return c.parseError(resp)
	}

	log.Printf("WorkOS: Updated organization %s", workosOrgID)
	return nil
}

// DeleteOrganization deletes an organization from WorkOS
func (c *Client) DeleteOrganization(workosOrgID string) error {
	if !c.Enabled() {
		return fmt.Errorf("workos client is not enabled")
	}

	httpReq, err := http.NewRequest("DELETE", c.baseURL+"/organizations/"+workosOrgID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode != 404 {
		return c.parseError(resp)
	}

	log.Printf("WorkOS: Deleted organization %s", workosOrgID)
	return nil
}

// ListOrganizations lists organizations from WorkOS with pagination
func (c *Client) ListOrganizations(limit int, after string) ([]Organization, string, error) {
	if !c.Enabled() {
		return nil, "", fmt.Errorf("workos client is not enabled")
	}

	// Build query parameters
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if after != "" {
		params.Set("after", after)
	}

	reqURL := c.baseURL + "/organizations"
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	httpReq, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, "", c.parseError(resp)
	}

	var listResp ListOrganizationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	return listResp.Data, listResp.ListMeta.After, nil
}

// VerifyWebhookSignature verifies a WorkOS webhook signature
func (c *Client) VerifyWebhookSignature(payload []byte, signature, timestamp string) bool {
	if c.config.WebhookSecret == "" {
		log.Println("WorkOS: Webhook secret not configured, skipping signature verification")
		return false
	}

	// WorkOS uses HMAC-SHA256 for webhook signatures
	// The signature format is: v1,<timestamp>,<payload>
	signedPayload := fmt.Sprintf("%s.%s", timestamp, string(payload))

	mac := hmac.New(sha256.New, []byte(c.config.WebhookSecret))
	mac.Write([]byte(signedPayload))
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// ParseWebhookEvent parses a webhook event from the request body
func (c *Client) ParseWebhookEvent(body []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook event: %w", err)
	}
	return &event, nil
}

// setHeaders sets the required headers for WorkOS API requests
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "prism-server/1.0")
}

// parseError parses an error response from WorkOS
func (c *Client) parseError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("workos api error: status %d", resp.StatusCode)
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("workos api error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return fmt.Errorf("workos api error: %s (code: %s)", errResp.Message, errResp.Code)
}
