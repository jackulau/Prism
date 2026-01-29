---
id: perplexity-provider
name: Add Perplexity AI Provider (Search-Augmented)
wave: 1
priority: 2
dependencies: []
estimated_hours: 3
tags:
- backend
- llm-provider
- search
- open-source
---

## Objective

Implement Perplexity AI provider for search-augmented generation using Llama-based models.

## Context

Perplexity provides search-augmented models built on open source:
- Llama 3.1 Sonar (with web search)
- Llama 3.1 Sonar Large
- Llama 3.1 Sonar Huge (405B with search)
- Returns citations with responses

OpenAI-compatible API with search grounding.

## Implementation

1. Create `/backend/internal/llm/perplexity/client.go`:
   - Implement the `Provider` interface
   - Use OpenAI-compatible API format
   - Support streaming responses
   - Handle citations in responses
   - Parse and include source URLs

2. Create `/backend/internal/llm/perplexity/models.go`:
   - Define Sonar model variants
   - Mark search capability

3. Register provider in `/backend/cmd/server/main.go`

## Acceptance Criteria

- [ ] Provider implements full `Provider` interface
- [ ] Streaming responses work correctly
- [ ] Citations included in responses
- [ ] API key validation works
- [ ] Search grounding active
- [ ] No security vulnerabilities

## Files to Create/Modify

- `backend/internal/llm/perplexity/client.go` - Create Perplexity client
- `backend/internal/llm/perplexity/models.go` - Create model definitions
- `backend/cmd/server/main.go` - Register provider

## Integration Points

- **Provides**: Perplexity provider for LLM Manager
- **Consumes**: LLM Provider interface
- **Conflicts**: None - independent provider implementation

## API Reference

```
Base URL: https://api.perplexity.ai
Auth: Authorization: Bearer <API_KEY>
Format: OpenAI-compatible (chat completions)
```
