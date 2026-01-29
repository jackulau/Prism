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

// PostHogToolConfig holds configuration for PostHog query tools
type PostHogToolConfig struct {
	APIKey    string
	ProjectID string
	Host      string // defaults to https://app.posthog.com
}

// PostHogListErrorsTool lists recent errors
type PostHogListErrorsTool struct {
	config     PostHogToolConfig
	httpClient *http.Client
}

func NewPostHogListErrorsTool(config PostHogToolConfig) *PostHogListErrorsTool {
	if config.Host == "" {
		config.Host = "https://app.posthog.com"
	}
	return &PostHogListErrorsTool{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *PostHogListErrorsTool) Name() string {
	return "posthog_list_errors"
}

func (t *PostHogListErrorsTool) Description() string {
	return "List recent errors and exceptions tracked in PostHog. Returns error messages, counts, and affected users."
}

func (t *PostHogListErrorsTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"limit": {
				Type:        "number",
				Description: "Maximum number of errors to return (default 20)",
			},
			"date_from": {
				Type:        "string",
				Description: "Start date for error search (ISO format, e.g., '2024-01-01')",
			},
			"date_to": {
				Type:        "string",
				Description: "End date for error search (ISO format)",
			},
			"search": {
				Type:        "string",
				Description: "Search term to filter errors",
			},
		},
		Required: []string{},
	}
}

func (t *PostHogListErrorsTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	limit := 20
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	endpoint := fmt.Sprintf("%s/api/projects/%s/query/", t.config.Host, t.config.ProjectID)

	query := fmt.Sprintf(`
		SELECT
			properties.$exception_message as message,
			properties.$exception_type as type,
			count() as count,
			max(timestamp) as last_seen,
			min(timestamp) as first_seen
		FROM events
		WHERE event = '$exception'
		%s
		GROUP BY message, type
		ORDER BY count DESC
		LIMIT %d
	`, buildDateFilter(params), limit)

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
		"errors":  result,
	}, nil
}

func (t *PostHogListErrorsTool) RequiresConfirmation() bool {
	return false
}

// buildDateFilter builds a date filter for HogQL
func buildDateFilter(params map[string]interface{}) string {
	filter := ""
	if dateFrom, ok := params["date_from"].(string); ok && dateFrom != "" {
		filter += fmt.Sprintf(" AND timestamp >= '%s'", dateFrom)
	}
	if dateTo, ok := params["date_to"].(string); ok && dateTo != "" {
		filter += fmt.Sprintf(" AND timestamp <= '%s'", dateTo)
	}
	if search, ok := params["search"].(string); ok && search != "" {
		filter += fmt.Sprintf(" AND properties.$exception_message ILIKE '%%%s%%'", search)
	}
	return filter
}

// PostHogErrorDetailsTool gets details of a specific error
type PostHogErrorDetailsTool struct {
	config     PostHogToolConfig
	httpClient *http.Client
}

func NewPostHogErrorDetailsTool(config PostHogToolConfig) *PostHogErrorDetailsTool {
	if config.Host == "" {
		config.Host = "https://app.posthog.com"
	}
	return &PostHogErrorDetailsTool{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *PostHogErrorDetailsTool) Name() string {
	return "posthog_error_details"
}

func (t *PostHogErrorDetailsTool) Description() string {
	return "Get detailed information about a specific error, including stack trace, affected users, and occurrence timeline."
}

func (t *PostHogErrorDetailsTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"error_message": {
				Type:        "string",
				Description: "The error message to look up",
			},
			"error_type": {
				Type:        "string",
				Description: "The error type/class (optional)",
			},
			"limit": {
				Type:        "number",
				Description: "Number of error occurrences to return (default 10)",
			},
		},
		Required: []string{"error_message"},
	}
}

func (t *PostHogErrorDetailsTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	errorMessage, ok := params["error_message"].(string)
	if !ok || errorMessage == "" {
		return nil, fmt.Errorf("error_message parameter is required")
	}

	limit := 10
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	endpoint := fmt.Sprintf("%s/api/projects/%s/query/", t.config.Host, t.config.ProjectID)

	// Build filter for error type if provided
	typeFilter := ""
	if errorType, ok := params["error_type"].(string); ok && errorType != "" {
		typeFilter = fmt.Sprintf(" AND properties.$exception_type = '%s'", errorType)
	}

	query := fmt.Sprintf(`
		SELECT
			timestamp,
			distinct_id as user_id,
			properties.$exception_message as message,
			properties.$exception_type as type,
			properties.$exception_stack_trace_raw as stack_trace,
			properties.$browser as browser,
			properties.$os as os,
			properties.$current_url as url
		FROM events
		WHERE event = '$exception'
		AND properties.$exception_message = '%s'
		%s
		ORDER BY timestamp DESC
		LIMIT %d
	`, errorMessage, typeFilter, limit)

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
		"success":       true,
		"error_message": errorMessage,
		"occurrences":   result,
	}, nil
}

func (t *PostHogErrorDetailsTool) RequiresConfirmation() bool {
	return false
}
