---
id: openai-llm-setup
name: OpenAI LLM Setup Enhancement
wave: 1
priority: 1
dependencies: []
estimated_hours: 3
tags:
- backend
- llm
- openai
---

## Objective

Enhance the OpenAI LLM integration with advanced features including function calling improvements, vision support, structured outputs, and model configuration options.

## Context

The codebase has an existing OpenAI client in `backend/internal/llm/openai/client.go` with:
- Basic chat completion streaming
- Tool/function calling support
- Model definitions (GPT-4, o3, o4-mini, etc.)
- API key validation

We need to enhance this with:
- JSON mode / structured outputs
- Vision (image) support
- Improved tool calling with parallel functions
- Model-specific parameters (temperature, top_p, etc.)
- Token counting and usage tracking
- Rate limiting and retry logic

## Implementation

### 1. Add Structured Output Support

**File**: `backend/internal/llm/openai/structured.go`

```go
package openai

type ResponseFormat struct {
    Type       string      `json:"type"` // "text", "json_object", "json_schema"
    JSONSchema *JSONSchema `json:"json_schema,omitempty"`
}

type JSONSchema struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description,omitempty"`
    Schema      map[string]interface{} `json:"schema"`
    Strict      bool                   `json:"strict,omitempty"`
}

func (c *Client) ChatWithSchema(ctx context.Context, req *llm.ChatRequest, schema *JSONSchema) (<-chan llm.StreamChunk, error)
```

### 2. Add Vision Support

**File**: `backend/internal/llm/openai/vision.go`

```go
package openai

type ImageContent struct {
    Type     string    `json:"type"` // "image_url"
    ImageURL *ImageURL `json:"image_url"`
}

type ImageURL struct {
    URL    string `json:"url"`    // base64 data URL or HTTP URL
    Detail string `json:"detail"` // "low", "high", "auto"
}

func (c *Client) formatVisionMessage(msg *llm.Message) (map[string]interface{}, error) {
    // Convert message with images to OpenAI format
    // Support both base64 and URL images
}
```

### 3. Enhance Chat Request Options

**File**: `backend/internal/llm/openai/options.go`

```go
package openai

type ChatOptions struct {
    Temperature      float64         `json:"temperature,omitempty"`
    TopP             float64         `json:"top_p,omitempty"`
    MaxTokens        int             `json:"max_tokens,omitempty"`
    PresencePenalty  float64         `json:"presence_penalty,omitempty"`
    FrequencyPenalty float64         `json:"frequency_penalty,omitempty"`
    Stop             []string        `json:"stop,omitempty"`
    ResponseFormat   *ResponseFormat `json:"response_format,omitempty"`
    Seed             *int            `json:"seed,omitempty"`
    ParallelToolCalls *bool          `json:"parallel_tool_calls,omitempty"`
}

func DefaultOptions() *ChatOptions
func (o *ChatOptions) Validate() error
func (o *ChatOptions) ApplyModelDefaults(model string) *ChatOptions
```

### 4. Add Token Counting

**File**: `backend/internal/llm/openai/tokens.go`

```go
package openai

import "github.com/pkoukk/tiktoken-go"

type TokenCounter struct {
    encoding *tiktoken.Tiktoken
}

func NewTokenCounter(model string) (*TokenCounter, error)
func (tc *TokenCounter) CountTokens(text string) int
func (tc *TokenCounter) CountMessages(messages []llm.Message) int
func (tc *TokenCounter) CountToolCalls(tools []llm.ToolDefinition) int
func EstimateMaxTokens(model string, inputTokens int) int
```

### 5. Add Usage Tracking

**File**: `backend/internal/llm/openai/usage.go`

```go
package openai

type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
    CachedTokens     int `json:"cached_tokens,omitempty"` // For prompt caching
}

type UsageTracker struct {
    mu     sync.Mutex
    usage  map[string]*Usage // by model
}

func (ut *UsageTracker) Record(model string, usage *Usage)
func (ut *UsageTracker) GetUsage(model string) *Usage
func (ut *UsageTracker) GetTotalUsage() *Usage
func (ut *UsageTracker) Reset()
```

### 6. Add Rate Limiting and Retry

**File**: `backend/internal/llm/openai/ratelimit.go`

```go
package openai

type RateLimiter struct {
    requestsPerMinute int
    tokensPerMinute   int
    requestBucket     *rate.Limiter
    tokenBucket       *rate.Limiter
    mu                sync.Mutex
}

func NewRateLimiter(rpm, tpm int) *RateLimiter
func (rl *RateLimiter) Wait(ctx context.Context, tokens int) error
func (rl *RateLimiter) UpdateFromHeaders(headers http.Header)

type RetryConfig struct {
    MaxRetries    int
    InitialDelay  time.Duration
    MaxDelay      time.Duration
    BackoffFactor float64
}

func (c *Client) chatWithRetry(ctx context.Context, req *http.Request, cfg *RetryConfig) (*http.Response, error)
```

### 7. Update Client

**File**: `backend/internal/llm/openai/client.go`

Update the existing client to use new features:

```go
type Client struct {
    apiKey       string
    baseURL      string
    client       *http.Client
    tokenCounter *TokenCounter
    usageTracker *UsageTracker
    rateLimiter  *RateLimiter
    retryConfig  *RetryConfig
}

func NewClient(apiKey string, opts ...ClientOption) *Client
func (c *Client) Chat(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamChunk, error)
func (c *Client) ChatWithOptions(ctx context.Context, req *llm.ChatRequest, opts *ChatOptions) (<-chan llm.StreamChunk, error)
```

### 8. Add Model Configurations

**File**: `backend/internal/llm/openai/models.go`

Update model definitions with capabilities:

```go
package openai

type ModelConfig struct {
    ID            string
    Name          string
    ContextWindow int
    MaxOutput     int
    SupportsVision bool
    SupportsTools  bool
    SupportJSON    bool
    InputCost      float64 // per 1M tokens
    OutputCost     float64 // per 1M tokens
}

var ModelConfigs = map[string]ModelConfig{
    "gpt-4o": {
        ContextWindow: 128000,
        MaxOutput:     16384,
        SupportsVision: true,
        SupportsTools:  true,
        SupportJSON:    true,
    },
    // ... other models
}
```

## Acceptance Criteria

- [ ] JSON mode / structured outputs working
- [ ] Vision support with base64 and URL images
- [ ] Model-specific parameters (temperature, top_p, etc.)
- [ ] Token counting with tiktoken
- [ ] Usage tracking per model
- [ ] Rate limiting with token bucket
- [ ] Retry logic with exponential backoff
- [ ] Model configurations with capabilities
- [ ] Parallel tool calls support
- [ ] Unit tests for token counting

## Files to Create/Modify

- `backend/internal/llm/openai/structured.go` - Structured outputs
- `backend/internal/llm/openai/vision.go` - Vision support
- `backend/internal/llm/openai/options.go` - Chat options
- `backend/internal/llm/openai/tokens.go` - Token counting
- `backend/internal/llm/openai/usage.go` - Usage tracking
- `backend/internal/llm/openai/ratelimit.go` - Rate limiting
- `backend/internal/llm/openai/models.go` - Model configs
- `backend/internal/llm/openai/client.go` - Update client

## Integration Points

- **Provides**: Enhanced OpenAI client with full feature support
- **Provides**: Token counting for context management
- **Provides**: Usage tracking for cost monitoring
- **Consumes**: LLM provider interface
- **Conflicts**: None - extends existing client
