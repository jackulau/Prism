package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jacklau/prism/internal/llm"
)

const (
	posthogDocsTimeout = 30 * time.Second
)

// PostHogDocsSearchTool searches PostHog documentation
type PostHogDocsSearchTool struct {
	httpClient *http.Client
}

// NewPostHogDocsSearchTool creates a new PostHog documentation search tool
func NewPostHogDocsSearchTool() *PostHogDocsSearchTool {
	return &PostHogDocsSearchTool{
		httpClient: &http.Client{
			Timeout: posthogDocsTimeout,
		},
	}
}

func (t *PostHogDocsSearchTool) Name() string {
	return "posthog_docs_search"
}

func (t *PostHogDocsSearchTool) Description() string {
	return "Search PostHog documentation for help with features, HogQL syntax, integrations, SDKs, and more. Returns relevant documentation pages and snippets."
}

func (t *PostHogDocsSearchTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"query": {
				Type:        "string",
				Description: "Search query for PostHog documentation (e.g., 'HogQL date functions', 'React SDK setup')",
			},
			"limit": {
				Type:        "number",
				Description: "Maximum number of results to return (default 5, max 20)",
			},
		},
		Required: []string{"query"},
	}
}

func (t *PostHogDocsSearchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	limit := 5
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
		if limit < 1 {
			limit = 1
		}
		if limit > 20 {
			limit = 20
		}
	}

	// PostHog public docs search API
	endpoint := fmt.Sprintf("https://posthog.com/api/search?query=%s&limit=%d",
		url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Prism/1.0")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		// Fallback on network error
		return t.fallbackSearch(query)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try fallback on error
		return t.fallbackSearch(query)
	}

	var results interface{}
	if err := json.Unmarshal(body, &results); err != nil {
		// Return raw if not JSON
		return map[string]interface{}{
			"success": true,
			"query":   query,
			"raw":     string(body),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"query":   query,
		"results": results,
	}, nil
}

// fallbackSearch provides helpful documentation links when the main API fails
func (t *PostHogDocsSearchTool) fallbackSearch(query string) (interface{}, error) {
	// Common documentation sections with direct links
	docSections := []map[string]string{
		{
			"title":       "HogQL Documentation",
			"url":         "https://posthog.com/docs/hogql",
			"description": "Complete HogQL query language reference",
		},
		{
			"title":       "Product Analytics",
			"url":         "https://posthog.com/docs/product-analytics",
			"description": "Insights, funnels, retention, and trends",
		},
		{
			"title":       "SDKs & Libraries",
			"url":         "https://posthog.com/docs/libraries",
			"description": "JavaScript, Python, React, iOS, Android SDKs",
		},
		{
			"title":       "API Reference",
			"url":         "https://posthog.com/docs/api",
			"description": "REST API documentation",
		},
		{
			"title":       "Error Tracking",
			"url":         "https://posthog.com/docs/error-tracking",
			"description": "Capturing and analyzing exceptions",
		},
	}

	return map[string]interface{}{
		"success": true,
		"query":   query,
		"message": "Search API unavailable. Here are relevant documentation sections:",
		"docs":    docSections,
		"tip":     fmt.Sprintf("Visit https://posthog.com/docs and search for '%s'", query),
	}, nil
}

func (t *PostHogDocsSearchTool) RequiresConfirmation() bool {
	return false // Read-only documentation search
}
