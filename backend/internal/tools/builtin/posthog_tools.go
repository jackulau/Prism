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

// PostHogQueryRunTool executes HogQL queries against PostHog
type PostHogQueryRunTool struct {
	config     PostHogConfig
	httpClient *http.Client
}

// NewPostHogQueryRunTool creates a new PostHog query run tool
func NewPostHogQueryRunTool(config PostHogConfig) *PostHogQueryRunTool {
	if config.Host == "" {
		config.Host = "https://app.posthog.com"
	}
	return &PostHogQueryRunTool{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (t *PostHogQueryRunTool) Name() string {
	return "posthog_query_run"
}

func (t *PostHogQueryRunTool) Description() string {
	return "Execute a HogQL query against PostHog analytics. Returns query results with events, persons, or aggregated data. Use for analytics queries like counting events, analyzing user behavior, etc."
}

func (t *PostHogQueryRunTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"query": {
				Type:        "string",
				Description: "The HogQL query to execute (e.g., 'SELECT event, count() FROM events GROUP BY event ORDER BY count() DESC LIMIT 10')",
			},
		},
		Required: []string{"query"},
	}
}

func (t *PostHogQueryRunTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	// Build PostHog query API request
	endpoint := fmt.Sprintf("%s/api/projects/%s/query/", t.config.Host, t.config.ProjectID)

	reqBody := map[string]interface{}{
		"query": map[string]interface{}{
			"kind":  "HogQLQuery",
			"query": query,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("API request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("PostHog API error (status %d): %s", resp.StatusCode, string(body)),
		}, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"results": result,
	}, nil
}

func (t *PostHogQueryRunTool) RequiresConfirmation() bool {
	return false // Read-only query operation
}

// PostHogGenerateQueryTool generates HogQL queries from natural language questions
type PostHogGenerateQueryTool struct {
	config     PostHogConfig
	httpClient *http.Client
}

// NewPostHogGenerateQueryTool creates a new PostHog query generation tool
func NewPostHogGenerateQueryTool(config PostHogConfig) *PostHogGenerateQueryTool {
	if config.Host == "" {
		config.Host = "https://app.posthog.com"
	}
	return &PostHogGenerateQueryTool{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *PostHogGenerateQueryTool) Name() string {
	return "posthog_generate_hogql"
}

func (t *PostHogGenerateQueryTool) Description() string {
	return "Generate a HogQL query from a natural language question about your analytics data. Use this when you need to construct a query but don't know the exact HogQL syntax."
}

func (t *PostHogGenerateQueryTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"question": {
				Type:        "string",
				Description: "Natural language question (e.g., 'How many users signed up last week?', 'What are the most common events?')",
			},
		},
		Required: []string{"question"},
	}
}

func (t *PostHogGenerateQueryTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	question, ok := params["question"].(string)
	if !ok || question == "" {
		return nil, fmt.Errorf("question parameter is required")
	}

	// PostHog AI query generation endpoint
	endpoint := fmt.Sprintf("%s/api/projects/%s/query/", t.config.Host, t.config.ProjectID)

	reqBody := map[string]interface{}{
		"query": map[string]interface{}{
			"kind":   "HogQLQuery",
			"query":  "",
			"prompt": question,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("API request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("PostHog API error (status %d): %s", resp.StatusCode, string(body)),
		}, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return map[string]interface{}{
		"success":         true,
		"generated_query": result,
	}, nil
}

func (t *PostHogGenerateQueryTool) RequiresConfirmation() bool {
	return false
}
