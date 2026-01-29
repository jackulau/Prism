package vercel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL = "https://api.vercel.com"
	defaultTimeout = 30 * time.Second
)

// APIClient is a client for the Vercel API
type APIClient struct {
	baseURL    string
	token      string
	teamID     string
	httpClient *http.Client
}

// NewAPIClient creates a new Vercel API client
func NewAPIClient(token, teamID string) *APIClient {
	return &APIClient{
		baseURL: defaultBaseURL,
		token:   token,
		teamID:  teamID,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// WithHTTPClient sets a custom HTTP client
func (c *APIClient) WithHTTPClient(client *http.Client) *APIClient {
	c.httpClient = client
	return c
}

// WithBaseURL sets a custom base URL (useful for testing)
func (c *APIClient) WithBaseURL(baseURL string) *APIClient {
	c.baseURL = baseURL
	return c
}

// CreateDeploymentRequest represents a request to create a deployment
type CreateDeploymentRequest struct {
	Name       string                 `json:"name"`
	Files      []DeploymentFile       `json:"files"`
	ProjectSettings *ProjectSettings  `json:"projectSettings,omitempty"`
	Target     string                 `json:"target,omitempty"` // "production" or "preview"
	GitSource  *GitSource             `json:"gitSource,omitempty"`
}

// DeploymentFile represents a file in a deployment
type DeploymentFile struct {
	File string `json:"file"` // Path to the file
	Data string `json:"data"` // Base64-encoded content or plain text
}

// ProjectSettings contains project configuration
type ProjectSettings struct {
	Framework       string `json:"framework,omitempty"`
	BuildCommand    string `json:"buildCommand,omitempty"`
	OutputDirectory string `json:"outputDirectory,omitempty"`
	RootDirectory   string `json:"rootDirectory,omitempty"`
	NodeVersion     string `json:"nodeVersion,omitempty"`
}

// GitSource represents git-based deployment source
type GitSource struct {
	Type   string `json:"type"` // "github", "gitlab", "bitbucket"
	RepoID string `json:"repoId"`
	Ref    string `json:"ref"`
}

// Deployment represents a Vercel deployment
type Deployment struct {
	ID         string            `json:"id"`
	URL        string            `json:"url"`
	Name       string            `json:"name"`
	State      string            `json:"state"` // "QUEUED", "BUILDING", "READY", "ERROR", "CANCELED"
	ReadyState string            `json:"readyState"`
	CreatedAt  int64             `json:"createdAt"`
	BuildingAt int64             `json:"buildingAt,omitempty"`
	Ready      int64             `json:"ready,omitempty"`
	Target     string            `json:"target,omitempty"`
	AliasError *AliasError       `json:"aliasError,omitempty"`
	Meta       map[string]string `json:"meta,omitempty"`
}

// AliasError represents an alias assignment error
type AliasError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// LogEntry represents a build log entry from Vercel
type VercelLogEntry struct {
	Type      string `json:"type"`
	Created   int64  `json:"created"`
	Payload   LogPayload `json:"payload"`
}

// LogPayload contains the log message details
type LogPayload struct {
	DeploymentID string `json:"deploymentId"`
	Text         string `json:"text"`
	StatusCode   int    `json:"statusCode,omitempty"`
	Proxy        *ProxyInfo `json:"proxy,omitempty"`
}

// ProxyInfo contains proxy information from logs
type ProxyInfo struct {
	Timestamp int64  `json:"timestamp"`
	Region    string `json:"region"`
	Path      string `json:"path"`
}

// ErrorResponse represents an error response from the Vercel API
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// ListDeploymentsResponse represents the response from listing deployments
type ListDeploymentsResponse struct {
	Deployments []Deployment `json:"deployments"`
	Pagination  Pagination   `json:"pagination"`
}

// Pagination contains pagination info
type Pagination struct {
	Count int    `json:"count"`
	Next  int64  `json:"next,omitempty"`
	Prev  int64  `json:"prev,omitempty"`
}

// doRequest performs an HTTP request with authentication
func (c *APIClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	// Build URL with team ID if present
	reqURL := c.baseURL + path
	if c.teamID != "" {
		parsedURL, err := url.Parse(reqURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}
		q := parsedURL.Query()
		q.Set("teamId", c.teamID)
		parsedURL.RawQuery = q.Encode()
		reqURL = parsedURL.String()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

// parseResponse parses the response body into the target or returns an error
func parseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return fmt.Errorf("API error (%s): %s", errResp.Error.Code, errResp.Error.Message)
		}
		return fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	if target != nil {
		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// CreateDeployment creates a new deployment
func (c *APIClient) CreateDeployment(ctx context.Context, req *CreateDeploymentRequest) (*Deployment, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/v13/deployments", req)
	if err != nil {
		return nil, err
	}

	var deployment Deployment
	if err := parseResponse(resp, &deployment); err != nil {
		return nil, err
	}

	return &deployment, nil
}

// GetDeployment retrieves a deployment by ID
func (c *APIClient) GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/v13/deployments/"+deploymentID, nil)
	if err != nil {
		return nil, err
	}

	var deployment Deployment
	if err := parseResponse(resp, &deployment); err != nil {
		return nil, err
	}

	return &deployment, nil
}

// ListDeployments lists deployments for a project
func (c *APIClient) ListDeployments(ctx context.Context, projectID string, limit int) ([]Deployment, error) {
	path := fmt.Sprintf("/v6/deployments?projectId=%s&limit=%d", url.QueryEscape(projectID), limit)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var listResp ListDeploymentsResponse
	if err := parseResponse(resp, &listResp); err != nil {
		return nil, err
	}

	return listResp.Deployments, nil
}

// GetBuildLogs retrieves build logs for a deployment
func (c *APIClient) GetBuildLogs(ctx context.Context, deploymentID string) ([]VercelLogEntry, error) {
	path := fmt.Sprintf("/v2/deployments/%s/events", deploymentID)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var logs []VercelLogEntry
	if err := parseResponse(resp, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// DeleteDeployment deletes a deployment
func (c *APIClient) DeleteDeployment(ctx context.Context, deploymentID string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/v13/deployments/"+deploymentID, nil)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}

// CancelDeployment cancels a deployment
func (c *APIClient) CancelDeployment(ctx context.Context, deploymentID string) (*Deployment, error) {
	resp, err := c.doRequest(ctx, http.MethodPatch, "/v12/deployments/"+deploymentID+"/cancel", nil)
	if err != nil {
		return nil, err
	}

	var deployment Deployment
	if err := parseResponse(resp, &deployment); err != nil {
		return nil, err
	}

	return &deployment, nil
}

// FrameworkToVercel converts our framework type to Vercel's framework identifier
func FrameworkToVercel(framework string) string {
	switch framework {
	case "nextjs":
		return "nextjs"
	case "react":
		return "create-react-app"
	case "vue":
		return "vue"
	case "vite":
		return "vite"
	case "static":
		return ""
	default:
		return ""
	}
}

// StateToDeploymentStatus converts Vercel state to our deployment status
func StateToDeploymentStatus(state string) string {
	switch state {
	case "QUEUED":
		return "queued"
	case "BUILDING":
		return "building"
	case "READY":
		return "ready"
	case "ERROR":
		return "error"
	case "CANCELED":
		return "cancelled"
	default:
		return "unknown"
	}
}
