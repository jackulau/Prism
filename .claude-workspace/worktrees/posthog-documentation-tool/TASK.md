---
id: posthog-documentation-tool
name: PostHog Tools - Documentation (posthog/documentation)
wave: 1
priority: 1
dependencies: []
estimated_hours: 2
tags:
- backend
- tools
- posthog
- mcp
---

## Objective

Implement the PostHog Documentation MCP tool for searching PostHog documentation.

## Context

This is a standalone tool that searches PostHog's public documentation. It doesn't require authentication and can be implemented independently.

The tool provides:
- `docs_search` - Search PostHog documentation for help with features, HogQL syntax, integrations, etc.

## Implementation

### 1. Create Documentation tool

**File:** `backend/internal/tools/builtin/posthog_docs.go`

```go
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

// PostHogDocsSearchTool searches PostHog documentation
type PostHogDocsSearchTool struct {
	httpClient *http.Client
}

func NewPostHogDocsSearchTool() *PostHogDocsSearchTool {
	return &PostHogDocsSearchTool{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
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
				Description: "Maximum number of results to return (default 5)",
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

	// PostHog uses Algolia for docs search
	// Public search endpoint
	endpoint := fmt.Sprintf("https://posthog.com/api/search?query=%s&limit=%d",
		url.QueryEscape(query), limit)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		// Fallback: try alternative approach with web fetch
		return t.fallbackSearch(ctx, query, limit)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try fallback on error
		return t.fallbackSearch(ctx, query, limit)
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

// fallbackSearch provides results when the main API fails
func (t *PostHogDocsSearchTool) fallbackSearch(ctx context.Context, query string, limit int) (interface{}, error) {
	// Generate helpful documentation links based on common query patterns
	docSections := []map[string]string{
		{
			"title": "HogQL Documentation",
			"url":   "https://posthog.com/docs/hogql",
			"description": "Complete HogQL query language reference",
		},
		{
			"title": "Product Analytics",
			"url":   "https://posthog.com/docs/product-analytics",
			"description": "Insights, funnels, retention, and trends",
		},
		{
			"title": "SDKs & Libraries",
			"url":   "https://posthog.com/docs/libraries",
			"description": "JavaScript, Python, React, iOS, Android SDKs",
		},
		{
			"title": "API Reference",
			"url":   "https://posthog.com/docs/api",
			"description": "REST API documentation",
		},
		{
			"title": "Error Tracking",
			"url":   "https://posthog.com/docs/error-tracking",
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
```

### 2. Register in init.go

This tool doesn't need PostHog API key - it's public documentation search:

```go
// PostHog Documentation search (no API key required)
if err := registry.Register(NewPostHogDocsSearchTool()); err != nil {
    return err
}
```

### 3. Add to read-only tools

In `approval.go`:
```go
"posthog_docs_search": true,
```

## Acceptance Criteria

- [ ] `posthog_docs_search` searches PostHog documentation
- [ ] Works without PostHog API key (public docs)
- [ ] Returns relevant documentation links and snippets
- [ ] Handles search API failures gracefully with fallback
- [ ] Respects limit parameter
- [ ] Marked as read-only (no confirmation required)

## Files to Create/Modify

- `backend/internal/tools/builtin/posthog_docs.go` - **CREATE** - Documentation search tool
- `backend/internal/tools/builtin/init.go` - **MODIFY** - Register the tool
- `backend/internal/tools/approval.go` - **MODIFY** - Add to read-only tools

## Integration Points

- **Provides**: PostHog documentation search
- **Consumes**: PostHog public docs API
- **Conflicts**: None - standalone implementation
