---
id: cloudprovider-anthropic
name: Anthropic Cloud Provider Implementation
wave: 2
priority: 2
dependencies:
- cloudprovider-interface
estimated_hours: 4
tags:
- backend
- anthropic
- implementation
---

## Objective

Implement the `CloudProvider` interface for Anthropic's cloud-based agent service (Claude.ai/API agents).

## Context

This provider will connect to Anthropic's agent API to create and manage cloud-hosted Claude agents. It follows the same pattern as the existing `backend/internal/llm/anthropic/` implementation but for agent lifecycle management rather than direct chat.

## Implementation

1. **Create package**: `backend/internal/cloudprovider/anthropic/`

2. **Create client file**: `backend/internal/cloudprovider/anthropic/client.go`
   ```go
   package anthropic
   
   import (
       "context"
       "net/http"
       
       "github.com/jacklau/prism/internal/cloudprovider"
   )
   
   type Client struct {
       apiKey     string
       httpClient *http.Client
       baseURL    string
   }
   
   func NewClient(apiKey string) *Client {
       return &Client{
           apiKey:     apiKey,
           httpClient: &http.Client{},
           baseURL:    "https://api.anthropic.com/v1",
       }
   }
   
   // Implement CloudProvider interface methods
   func (c *Client) Name() string { return "anthropic-cloud" }
   
   func (c *Client) CreateAgent(ctx context.Context, params cloudprovider.CreateAgentParams) (*cloudprovider.Agent, error) {
       // API call to create agent
   }
   
   func (c *Client) GetAgent(ctx context.Context, agentID string) (*cloudprovider.Agent, error) {
       // API call to get agent
   }
   
   func (c *Client) DeleteAgent(ctx context.Context, agentID string) error {
       // API call to delete agent
   }
   
   func (c *Client) GetMessages(ctx context.Context, agent *cloudprovider.Agent) ([]cloudprovider.ProviderMessage, error) {
       // API call to get messages
   }
   
   func (c *Client) SendMessage(ctx context.Context, agent *cloudprovider.Agent, message string, images []cloudprovider.ImageData) (bool, error) {
       // API call to send message
   }
   
   func (c *Client) StreamMessages(ctx context.Context, agent *cloudprovider.Agent) (<-chan cloudprovider.MessageChunk, error) {
       // SSE streaming from API
   }
   
   func (c *Client) ValidateCredentials(ctx context.Context) error {
       // Validate API key
   }
   
   func (c *Client) HasCredentials() bool {
       return c.apiKey != ""
   }
   ```

3. **Create API types**: `backend/internal/cloudprovider/anthropic/types.go`
   - Request/response structs matching Anthropic's API schema
   - Type conversion helpers between API types and cloudprovider types

4. **Create streaming handler**: `backend/internal/cloudprovider/anthropic/stream.go`
   - SSE parsing for streaming responses
   - Chunk processing and channel management

## Acceptance Criteria

- [ ] Client implements all `CloudProvider` interface methods
- [ ] API requests include proper authentication headers
- [ ] Streaming responses are properly parsed and channeled
- [ ] Error responses are converted to appropriate cloudprovider errors
- [ ] Rate limiting is handled gracefully
- [ ] Context cancellation is properly propagated

## Files to Create/Modify

- `backend/internal/cloudprovider/anthropic/client.go` - Main client
- `backend/internal/cloudprovider/anthropic/types.go` - API types
- `backend/internal/cloudprovider/anthropic/stream.go` - Streaming handler

## Integration Points

- **Provides**: Anthropic CloudProvider implementation
- **Consumes**: cloudprovider.CloudProvider interface
- **Conflicts**: None - new package
