---
id: posthog-insights-tools
name: PostHog Tools - Insights (posthog/insights)
wave: 2
priority: 2
dependencies:
- posthog-query-runner
estimated_hours: 5
tags:
- backend
- tools
- posthog
- mcp
---

## Objective

Implement the PostHog Insights MCP tool set with CRUD operations for PostHog insights.

## Context

PostHog Insights are saved visualizations and queries. This tool set provides:
- `insight_create_from_query` - Create new insight from a query
- `insight_delete` - Delete an insight
- `insight_get` - Get insight details
- `insight_query` - Query insight data
- `insight_update` - Update insight configuration
- `insights_get_all` - List all insights

## Implementation

### 1. Create Insights tools file

**File:** `backend/internal/tools/builtin/posthog_insights.go`

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

// PostHogInsightsGetAllTool lists all insights
type PostHogInsightsGetAllTool struct {
	config     PostHogConfig
	httpClient *http.Client
}

func NewPostHogInsightsGetAllTool(config PostHogConfig) *PostHogInsightsGetAllTool {
	if config.Host == "" {
		config.Host = "https://app.posthog.com"
	}
	return &PostHogInsightsGetAllTool{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *PostHogInsightsGetAllTool) Name() string {
	return "posthog_insights_get_all"
}

func (t *PostHogInsightsGetAllTool) Description() string {
	return "List all saved insights in your PostHog project. Returns insight IDs, names, and types."
}

func (t *PostHogInsightsGetAllTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"limit": {
				Type:        "number",
				Description: "Maximum number of insights to return (default 100)",
			},
			"saved": {
				Type:        "boolean",
				Description: "Only return saved insights (default true)",
			},
		},
		Required: []string{},
	}
}

func (t *PostHogInsightsGetAllTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	limit := 100
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	saved := true
	if s, ok := params["saved"].(bool); ok {
		saved = s
	}

	endpoint := fmt.Sprintf("%s/api/projects/%s/insights/?limit=%d&saved=%t",
		t.config.Host, t.config.ProjectID, limit, saved)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.config.APIKey)

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
		"success":  true,
		"insights": result,
	}, nil
}

func (t *PostHogInsightsGetAllTool) RequiresConfirmation() bool {
	return false
}

// PostHogInsightGetTool gets a specific insight
type PostHogInsightGetTool struct {
	config     PostHogConfig
	httpClient *http.Client
}

func NewPostHogInsightGetTool(config PostHogConfig) *PostHogInsightGetTool {
	if config.Host == "" {
		config.Host = "https://app.posthog.com"
	}
	return &PostHogInsightGetTool{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *PostHogInsightGetTool) Name() string {
	return "posthog_insight_get"
}

func (t *PostHogInsightGetTool) Description() string {
	return "Get details of a specific PostHog insight by ID or short ID."
}

func (t *PostHogInsightGetTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"insight_id": {
				Type:        "string",
				Description: "The insight ID or short ID",
			},
		},
		Required: []string{"insight_id"},
	}
}

func (t *PostHogInsightGetTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	insightID, ok := params["insight_id"].(string)
	if !ok || insightID == "" {
		return nil, fmt.Errorf("insight_id parameter is required")
	}

	endpoint := fmt.Sprintf("%s/api/projects/%s/insights/%s/",
		t.config.Host, t.config.ProjectID, insightID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.config.APIKey)

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
		"insight": result,
	}, nil
}

func (t *PostHogInsightGetTool) RequiresConfirmation() bool {
	return false
}

// PostHogInsightQueryTool refreshes and queries an insight
type PostHogInsightQueryTool struct {
	config     PostHogConfig
	httpClient *http.Client
}

func NewPostHogInsightQueryTool(config PostHogConfig) *PostHogInsightQueryTool {
	if config.Host == "" {
		config.Host = "https://app.posthog.com"
	}
	return &PostHogInsightQueryTool{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (t *PostHogInsightQueryTool) Name() string {
	return "posthog_insight_query"
}

func (t *PostHogInsightQueryTool) Description() string {
	return "Execute and refresh an insight query, returning the latest data."
}

func (t *PostHogInsightQueryTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"insight_id": {
				Type:        "string",
				Description: "The insight ID to query",
			},
			"refresh": {
				Type:        "boolean",
				Description: "Force refresh the data (default false)",
			},
		},
		Required: []string{"insight_id"},
	}
}

func (t *PostHogInsightQueryTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	insightID, ok := params["insight_id"].(string)
	if !ok || insightID == "" {
		return nil, fmt.Errorf("insight_id parameter is required")
	}

	refresh := false
	if r, ok := params["refresh"].(bool); ok {
		refresh = r
	}

	endpoint := fmt.Sprintf("%s/api/projects/%s/insights/%s/?refresh=%t",
		t.config.Host, t.config.ProjectID, insightID, refresh)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.config.APIKey)

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
		"data":    result,
	}, nil
}

func (t *PostHogInsightQueryTool) RequiresConfirmation() bool {
	return false
}

// PostHogInsightCreateTool creates a new insight
type PostHogInsightCreateTool struct {
	config     PostHogConfig
	httpClient *http.Client
}

func NewPostHogInsightCreateTool(config PostHogConfig) *PostHogInsightCreateTool {
	if config.Host == "" {
		config.Host = "https://app.posthog.com"
	}
	return &PostHogInsightCreateTool{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *PostHogInsightCreateTool) Name() string {
	return "posthog_insight_create"
}

func (t *PostHogInsightCreateTool) Description() string {
	return "Create a new PostHog insight from a query definition."
}

func (t *PostHogInsightCreateTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"name": {
				Type:        "string",
				Description: "Name for the new insight",
			},
			"query": {
				Type:        "string",
				Description: "HogQL query or query definition JSON",
			},
			"description": {
				Type:        "string",
				Description: "Optional description for the insight",
			},
		},
		Required: []string{"name", "query"},
	}
}

func (t *PostHogInsightCreateTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name parameter is required")
	}

	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	endpoint := fmt.Sprintf("%s/api/projects/%s/insights/",
		t.config.Host, t.config.ProjectID)

	reqBody := map[string]interface{}{
		"name":  name,
		"query": map[string]interface{}{
			"kind":  "HogQLQuery",
			"query": query,
		},
		"saved": true,
	}

	if desc, ok := params["description"].(string); ok && desc != "" {
		reqBody["description"] = desc
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
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
		"insight": result,
	}, nil
}

func (t *PostHogInsightCreateTool) RequiresConfirmation() bool {
	return true // Creates data
}

// PostHogInsightUpdateTool updates an existing insight
type PostHogInsightUpdateTool struct {
	config     PostHogConfig
	httpClient *http.Client
}

func NewPostHogInsightUpdateTool(config PostHogConfig) *PostHogInsightUpdateTool {
	if config.Host == "" {
		config.Host = "https://app.posthog.com"
	}
	return &PostHogInsightUpdateTool{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *PostHogInsightUpdateTool) Name() string {
	return "posthog_insight_update"
}

func (t *PostHogInsightUpdateTool) Description() string {
	return "Update an existing PostHog insight's name, description, or query."
}

func (t *PostHogInsightUpdateTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"insight_id": {
				Type:        "string",
				Description: "The insight ID to update",
			},
			"name": {
				Type:        "string",
				Description: "New name for the insight",
			},
			"description": {
				Type:        "string",
				Description: "New description for the insight",
			},
			"query": {
				Type:        "string",
				Description: "New HogQL query",
			},
		},
		Required: []string{"insight_id"},
	}
}

func (t *PostHogInsightUpdateTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	insightID, ok := params["insight_id"].(string)
	if !ok || insightID == "" {
		return nil, fmt.Errorf("insight_id parameter is required")
	}

	endpoint := fmt.Sprintf("%s/api/projects/%s/insights/%s/",
		t.config.Host, t.config.ProjectID, insightID)

	reqBody := map[string]interface{}{}

	if name, ok := params["name"].(string); ok && name != "" {
		reqBody["name"] = name
	}
	if desc, ok := params["description"].(string); ok {
		reqBody["description"] = desc
	}
	if query, ok := params["query"].(string); ok && query != "" {
		reqBody["query"] = map[string]interface{}{
			"kind":  "HogQLQuery",
			"query": query,
		}
	}

	if len(reqBody) == 0 {
		return nil, fmt.Errorf("at least one field to update is required")
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", endpoint, bytes.NewBuffer(jsonBody))
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
		"insight": result,
	}, nil
}

func (t *PostHogInsightUpdateTool) RequiresConfirmation() bool {
	return true // Modifies data
}

// PostHogInsightDeleteTool deletes an insight
type PostHogInsightDeleteTool struct {
	config     PostHogConfig
	httpClient *http.Client
}

func NewPostHogInsightDeleteTool(config PostHogConfig) *PostHogInsightDeleteTool {
	if config.Host == "" {
		config.Host = "https://app.posthog.com"
	}
	return &PostHogInsightDeleteTool{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *PostHogInsightDeleteTool) Name() string {
	return "posthog_insight_delete"
}

func (t *PostHogInsightDeleteTool) Description() string {
	return "Delete a PostHog insight by ID."
}

func (t *PostHogInsightDeleteTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"insight_id": {
				Type:        "string",
				Description: "The insight ID to delete",
			},
		},
		Required: []string{"insight_id"},
	}
}

func (t *PostHogInsightDeleteTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	insightID, ok := params["insight_id"].(string)
	if !ok || insightID == "" {
		return nil, fmt.Errorf("insight_id parameter is required")
	}

	endpoint := fmt.Sprintf("%s/api/projects/%s/insights/%s/",
		t.config.Host, t.config.ProjectID, insightID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.config.APIKey)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("API request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("PostHog API error (status %d): %s", resp.StatusCode, string(body)),
		}, nil
	}

	return map[string]interface{}{
		"success":    true,
		"deleted_id": insightID,
	}, nil
}

func (t *PostHogInsightDeleteTool) RequiresConfirmation() bool {
	return true // Deletes data
}
```

### 2. Register in init.go

Add registration for all insight tools:

```go
// PostHog Insights tools (require API key)
if config.PostHogConfig != nil && config.PostHogConfig.APIKey != "" {
    if err := registry.Register(NewPostHogInsightsGetAllTool(*config.PostHogConfig)); err != nil {
        return err
    }
    if err := registry.Register(NewPostHogInsightGetTool(*config.PostHogConfig)); err != nil {
        return err
    }
    if err := registry.Register(NewPostHogInsightQueryTool(*config.PostHogConfig)); err != nil {
        return err
    }
    if err := registry.Register(NewPostHogInsightCreateTool(*config.PostHogConfig)); err != nil {
        return err
    }
    if err := registry.Register(NewPostHogInsightUpdateTool(*config.PostHogConfig)); err != nil {
        return err
    }
    if err := registry.Register(NewPostHogInsightDeleteTool(*config.PostHogConfig)); err != nil {
        return err
    }
}
```

## Acceptance Criteria

- [ ] `posthog_insights_get_all` lists all saved insights
- [ ] `posthog_insight_get` retrieves insight by ID
- [ ] `posthog_insight_query` executes insight query with refresh option
- [ ] `posthog_insight_create` creates new insight from query
- [ ] `posthog_insight_update` updates insight properties
- [ ] `posthog_insight_delete` deletes insight by ID
- [ ] Read operations don't require confirmation
- [ ] Write operations (create, update, delete) require confirmation
- [ ] All tools handle API errors gracefully

## Files to Create/Modify

- `backend/internal/tools/builtin/posthog_insights.go` - **CREATE** - Insights tool implementations
- `backend/internal/tools/builtin/init.go` - **MODIFY** - Register insight tools
- `backend/internal/tools/approval.go` - **MODIFY** - Add read-only tools

## Integration Points

- **Provides**: PostHog insight management tools
- **Consumes**: PostHog API, PostHogConfig from posthog_tools.go
- **Conflicts**: Depends on posthog-query-runner for shared config
