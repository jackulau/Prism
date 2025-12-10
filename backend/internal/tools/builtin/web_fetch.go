package builtin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/jacklau/prism/internal/llm"
)

const (
	webFetchTimeout    = 30 * time.Second
	maxContentLength   = 5 * 1024 * 1024 // 5MB max
	cacheTTL           = 15 * time.Minute
	maxCacheEntries    = 100
)

// WebFetchConfig holds configuration for the web fetch tool
type WebFetchConfig struct {
	LLMManager llm.Provider // Optional LLM provider for AI analysis
}

// cacheEntry represents a cached fetch result
type cacheEntry struct {
	content   string
	title     string
	fetchedAt time.Time
}

// WebFetchTool fetches and processes web content
type WebFetchTool struct {
	httpClient *http.Client
	llmManager llm.Provider
	cache      map[string]*cacheEntry
	cacheMu    sync.RWMutex
}

// NewWebFetchTool creates a new web fetch tool
func NewWebFetchTool(config WebFetchConfig) *WebFetchTool {
	return &WebFetchTool{
		httpClient: &http.Client{
			Timeout: webFetchTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Allow up to 10 redirects
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		llmManager: config.LLMManager,
		cache:      make(map[string]*cacheEntry),
	}
}

func (t *WebFetchTool) Name() string {
	return "web_fetch"
}

func (t *WebFetchTool) Description() string {
	return `Fetches content from a URL and converts HTML to markdown. Optionally processes the content with an AI prompt. Use this to retrieve and analyze web content. HTTP URLs are automatically upgraded to HTTPS.`
}

func (t *WebFetchTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"url": {
				Type:        "string",
				Description: "The URL to fetch content from (must be http or https)",
			},
			"prompt": {
				Type:        "string",
				Description: "Optional prompt for AI analysis of the fetched content",
			},
		},
		Required: []string{"url"},
	}
}

func (t *WebFetchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	urlStr, ok := params["url"].(string)
	if !ok || urlStr == "" {
		return nil, fmt.Errorf("url parameter is required")
	}

	prompt, _ := params["prompt"].(string)

	// Parse and validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow http and https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("URL must use http or https scheme")
	}

	// Upgrade HTTP to HTTPS
	if parsedURL.Scheme == "http" {
		parsedURL.Scheme = "https"
		urlStr = parsedURL.String()
	}

	// Check cache first
	if cached := t.getFromCache(urlStr); cached != nil {
		result := map[string]interface{}{
			"content": cached.content,
			"title":   cached.title,
			"url":     urlStr,
			"cached":  true,
		}

		// Process with AI if prompt provided
		if prompt != "" && t.llmManager != nil {
			analysis, err := t.analyzeContent(ctx, cached.content, prompt)
			if err == nil {
				result["analysis"] = analysis
			}
		}

		return result, nil
	}

	// Fetch content
	content, title, finalURL, err := t.fetchURL(ctx, urlStr)
	if err != nil {
		return nil, err
	}

	// Check for cross-host redirect
	originalHost := parsedURL.Host
	finalParsedURL, _ := url.Parse(finalURL)
	if finalParsedURL != nil && finalParsedURL.Host != originalHost {
		return map[string]interface{}{
			"redirect":     true,
			"redirect_url": finalURL,
			"message":      fmt.Sprintf("URL redirects to a different host: %s. Please make a new request with the redirect URL.", finalURL),
		}, nil
	}

	// Cache the result
	t.addToCache(urlStr, content, title)

	result := map[string]interface{}{
		"content": content,
		"title":   title,
		"url":     finalURL,
		"cached":  false,
	}

	// Process with AI if prompt provided
	if prompt != "" && t.llmManager != nil {
		analysis, err := t.analyzeContent(ctx, content, prompt)
		if err == nil {
			result["analysis"] = analysis
		}
	}

	return result, nil
}

func (t *WebFetchTool) fetchURL(ctx context.Context, urlStr string) (content, title, finalURL string, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to look like a browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Get final URL after redirects
	finalURL = resp.Request.URL.String()

	// Check content length
	if resp.ContentLength > maxContentLength {
		return "", "", "", fmt.Errorf("content too large: %d bytes (max %d)", resp.ContentLength, maxContentLength)
	}

	// Read body with limit
	limitedReader := io.LimitReader(resp.Body, maxContentLength)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml") {
		// Parse HTML and convert to markdown
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to parse HTML: %w", err)
		}

		// Extract title
		title = doc.Find("title").First().Text()
		title = strings.TrimSpace(title)

		// Remove script, style, and other non-content elements
		doc.Find("script, style, nav, footer, header, aside, .ads, .advertisement").Remove()

		// Get main content
		html, err := doc.Find("body").Html()
		if err != nil {
			html, _ = doc.Html()
		}

		// Convert to markdown
		converter := md.NewConverter("", true, nil)
		content, err = converter.ConvertString(html)
		if err != nil {
			// Fall back to plain text
			content = doc.Find("body").Text()
		}

		// Clean up excessive whitespace
		content = cleanupMarkdown(content)

	} else {
		// Return as plain text
		content = string(body)
	}

	return content, title, finalURL, nil
}

func cleanupMarkdown(content string) string {
	// Replace multiple newlines with double newline
	lines := strings.Split(content, "\n")
	var result []string
	prevEmpty := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isEmpty := trimmed == ""

		if isEmpty {
			if !prevEmpty {
				result = append(result, "")
			}
			prevEmpty = true
		} else {
			result = append(result, trimmed)
			prevEmpty = false
		}
	}

	return strings.Join(result, "\n")
}

func (t *WebFetchTool) analyzeContent(ctx context.Context, content, prompt string) (string, error) {
	if t.llmManager == nil {
		return "", fmt.Errorf("LLM manager not configured")
	}

	// Truncate content if too long
	maxContentForAnalysis := 50000
	if len(content) > maxContentForAnalysis {
		content = content[:maxContentForAnalysis] + "\n...[truncated]"
	}

	// Create chat request
	messages := []llm.Message{
		{
			Role:    "system",
			Content: "You are analyzing web content. Provide a concise, helpful response to the user's query about the content.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Content:\n%s\n\nQuery: %s", content, prompt),
		},
	}

	req := &llm.ChatRequest{
		Messages:    messages,
		Temperature: 0.3,
		MaxTokens:   1000,
		Stream:      false,
	}

	// Get response
	chunks, err := t.llmManager.Chat(ctx, req)
	if err != nil {
		return "", err
	}

	// Collect response
	var response strings.Builder
	for chunk := range chunks {
		if chunk.Error != nil {
			return "", chunk.Error
		}
		response.WriteString(chunk.Delta)
	}

	return response.String(), nil
}

func (t *WebFetchTool) getFromCache(url string) *cacheEntry {
	t.cacheMu.RLock()
	defer t.cacheMu.RUnlock()

	entry, ok := t.cache[url]
	if !ok {
		return nil
	}

	// Check if expired
	if time.Since(entry.fetchedAt) > cacheTTL {
		return nil
	}

	return entry
}

func (t *WebFetchTool) addToCache(url, content, title string) {
	t.cacheMu.Lock()
	defer t.cacheMu.Unlock()

	// Evict old entries if cache is full
	if len(t.cache) >= maxCacheEntries {
		// Remove oldest entry
		var oldestURL string
		var oldestTime time.Time
		for u, entry := range t.cache {
			if oldestURL == "" || entry.fetchedAt.Before(oldestTime) {
				oldestURL = u
				oldestTime = entry.fetchedAt
			}
		}
		if oldestURL != "" {
			delete(t.cache, oldestURL)
		}
	}

	t.cache[url] = &cacheEntry{
		content:   content,
		title:     title,
		fetchedAt: time.Now(),
	}
}

func (t *WebFetchTool) RequiresConfirmation() bool {
	return false
}
