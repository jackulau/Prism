---
id: posthog-query-runner
name: PostHog Tools - Query Runner (posthog/query_runner)
wave: 1
priority: 2
dependencies: []
estimated_hours: 5
tags:
- backend
- tools
- posthog
- mcp
---

## Objective

Implement the PostHog Query Runner MCP tool set with `query_run`, `query_generate_hogql_from_question`, and `docs_search` tools.

## Context

PostHog is an analytics platform. This MCP tool set enables AI assistants to:
1. Execute HogQL queries against PostHog
2. Generate HogQL queries from natural language questions
3. Search PostHog documentation

This will be implemented as built-in tools that connect to PostHog's API.

## Implementation

### 1. Create PostHog client

**File:** `backend/internal/tools/builtin/posthog_tools.go`

```go
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

// PostHogConfig holds PostHog API configuration
type PostHogConfig struct {
	APIKey     string
	ProjectID  string
	Host       string // defaults to https://app.posthog.com
}

// PostHogQueryRunTool executes HogQL queries
type PostHogQueryRunTool struct {
	config     PostHogConfig
	httpClient *http.Client
}

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
	return "Execute a HogQL query against PostHog analytics. Returns query results with events, persons, or aggregated data."
}

func (t *PostHogQueryRunTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"query": {
				Type:        "string",
				Description: "The HogQL query to execute (e.g., 'SELECT * FROM events LIMIT 10')",
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
```

### 2. Add Natural Language to HogQL tool

```go
// PostHogGenerateQueryTool generates HogQL from natural language
type PostHogGenerateQueryTool struct {
	config     PostHogConfig
	httpClient *http.Client
}

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
	return "Generate a HogQL query from a natural language question about your analytics data."
}

func (t *PostHogGenerateQueryTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"question": {
				Type:        "string",
				Description: "Natural language question (e.g., 'How many users signed up last week?')",
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
			"kind":     "HogQLQuery",
			"query":    "", // Empty - will be generated
			"prompt":   question,
			"aiGenerated": true,
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
```

### 3. Add PostHog docs search tool

```go
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
	return "Search PostHog documentation for help with features, HogQL syntax, integrations, and more."
}

func (t *PostHogDocsSearchTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"query": {
				Type:        "string",
				Description: "Search query for PostHog documentation",
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

	// PostHog docs search API
	endpoint := fmt.Sprintf("https://posthog.com/api/docs-search?query=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("docs search failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(body, &results); err != nil {
		// Return raw response if not JSON
		return map[string]interface{}{
			"success": true,
			"raw":     string(body),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"results": results,
	}, nil
}

func (t *PostHogDocsSearchTool) RequiresConfirmation() bool {
	return false
}
```

### 4. Add configuration and registration

**File:** `backend/internal/config/config.go`

Add PostHog configuration:
```go
// PostHog tools configuration
PostHogAPIKey    string `env:"POSTHOG_TOOLS_API_KEY"`
PostHogProjectID string `env:"POSTHOG_PROJECT_ID"`
PostHogHost      string `env:"POSTHOG_HOST" default:"https://app.posthog.com"`
```

**File:** `backend/internal/tools/builtin/init.go`

Add to Config struct and RegisterAll:
```go
// In Config struct
PostHogConfig *PostHogConfig

// In RegisterAll
if config.PostHogConfig != nil && config.PostHogConfig.APIKey != "" {
    if err := registry.Register(NewPostHogQueryRunTool(*config.PostHogConfig)); err != nil {
        return err
    }
    if err := registry.Register(NewPostHogGenerateQueryTool(*config.PostHogConfig)); err != nil {
        return err
    }
}
// Docs search doesn't need API key
if err := registry.Register(NewPostHogDocsSearchTool()); err != nil {
    return err
}
```

## Acceptance Criteria

- [ ] `posthog_query_run` executes HogQL queries against PostHog API
- [ ] `posthog_generate_hogql` generates queries from natural language
- [ ] `posthog_docs_search` searches PostHog documentation
- [ ] All tools return structured success/error responses
- [ ] Configuration via environment variables
- [ ] Proper error handling for API failures
- [ ] Timeout handling for long queries
- [ ] All marked as read-only (no confirmation required)

## Files to Create/Modify

- `backend/internal/tools/builtin/posthog_tools.go` - **CREATE** - PostHog tool implementations
- `backend/internal/tools/builtin/init.go` - **MODIFY** - Add PostHog config and registration
- `backend/internal/config/config.go` - **MODIFY** - Add PostHog environment variables
- `backend/internal/tools/approval.go` - **MODIFY** - Add tools to read-only list
- `backend/cmd/server/main.go` - **MODIFY** - Wire up PostHog config

## Integration Points

- **Provides**: PostHog query tools for analytics
- **Consumes**: PostHog API via HTTP client
- **Conflicts**: None - new file creation
