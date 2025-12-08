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

// SearchResult represents a single search result
type SearchResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// WebSearchTool searches the web for information
type WebSearchTool struct {
	serpAPIKey       string
	googleAPIKey     string
	googleSearchCX   string
	httpClient       *http.Client
}

// WebSearchConfig holds configuration for the web search tool
type WebSearchConfig struct {
	SerpAPIKey     string
	GoogleAPIKey   string
	GoogleSearchCX string
}

// NewWebSearchTool creates a new web search tool
func NewWebSearchTool(config WebSearchConfig) *WebSearchTool {
	return &WebSearchTool{
		serpAPIKey:     config.SerpAPIKey,
		googleAPIKey:   config.GoogleAPIKey,
		googleSearchCX: config.GoogleSearchCX,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *WebSearchTool) Name() string {
	return "web_search"
}

func (t *WebSearchTool) Description() string {
	return "Search the web for information. Returns a list of search results with titles, URLs, and snippets. Use this to find current information, documentation, or answers to questions."
}

func (t *WebSearchTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"query": {
				Type:        "string",
				Description: "The search query",
			},
			"num_results": {
				Type:        "integer",
				Description: "Number of results to return (1-10)",
				Default:     5,
			},
		},
		Required: []string{"query"},
	}
}

func (t *WebSearchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	numResults := 5
	if n, ok := params["num_results"].(float64); ok {
		numResults = int(n)
		if numResults < 1 {
			numResults = 1
		}
		if numResults > 10 {
			numResults = 10
		}
	}

	var results []SearchResult
	var err error

	// Try SerpAPI first, then Google Custom Search
	if t.serpAPIKey != "" {
		results, err = t.searchWithSerpAPI(ctx, query, numResults)
	} else if t.googleAPIKey != "" && t.googleSearchCX != "" {
		results, err = t.searchWithGoogle(ctx, query, numResults)
	} else {
		return nil, fmt.Errorf("no search API configured. Set SERPAPI_KEY or GOOGLE_SEARCH_API_KEY and GOOGLE_SEARCH_CX")
	}

	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return map[string]interface{}{
		"query":   query,
		"results": results,
		"count":   len(results),
	}, nil
}

func (t *WebSearchTool) RequiresConfirmation() bool {
	return false
}

func (t *WebSearchTool) searchWithSerpAPI(ctx context.Context, query string, numResults int) ([]SearchResult, error) {
	apiURL := fmt.Sprintf(
		"https://serpapi.com/search.json?q=%s&num=%d&api_key=%s",
		url.QueryEscape(query),
		numResults,
		t.serpAPIKey,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SerpAPI error: %s", string(body))
	}

	var serpResp struct {
		OrganicResults []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"organic_results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&serpResp); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(serpResp.OrganicResults))
	for _, r := range serpResp.OrganicResults {
		results = append(results, SearchResult{
			Title:   r.Title,
			Link:    r.Link,
			Snippet: r.Snippet,
		})
	}

	return results, nil
}

func (t *WebSearchTool) searchWithGoogle(ctx context.Context, query string, numResults int) ([]SearchResult, error) {
	apiURL := fmt.Sprintf(
		"https://www.googleapis.com/customsearch/v1?q=%s&num=%d&key=%s&cx=%s",
		url.QueryEscape(query),
		numResults,
		t.googleAPIKey,
		t.googleSearchCX,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Google Search API error: %s", string(body))
	}

	var googleResp struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(googleResp.Items))
	for _, item := range googleResp.Items {
		results = append(results, SearchResult{
			Title:   item.Title,
			Link:    item.Link,
			Snippet: item.Snippet,
		})
	}

	return results, nil
}
