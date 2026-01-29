---
id: openrouter-provider
name: Add OpenRouter Provider for 200+ Models
wave: 1
priority: 1
dependencies: []
estimated_hours: 4
tags:
- backend
- llm-provider
- open-source
---

## Objective

Implement OpenRouter provider to access 200+ open source and commercial models through a single API endpoint.

## Context

OpenRouter (openrouter.ai) provides unified access to models from many providers including:
- Meta Llama 3.1/3.2/3.3 (8B, 70B, 405B)
- Mistral/Mixtral models
- DeepSeek models
- Qwen models
- Google Gemma
- Anthropic Claude (fallback)
- OpenAI models (fallback)

This is the highest-impact single integration for open source model support.

## Implementation

1. Create `/backend/internal/llm/openrouter/client.go`:
   - Implement the `Provider` interface
   - Use OpenAI-compatible API format (OpenRouter is OpenAI-compatible)
   - Support streaming responses
   - Implement tool calling for compatible models
   - Handle model-specific context windows and capabilities

2. Create `/backend/internal/llm/openrouter/models.go`:
   - Define popular open source models with their capabilities
   - Include context windows, tool support, vision support flags
   - Dynamically fetch available models from OpenRouter API

3. Register provider in `/backend/cmd/server/main.go`:
   - Add OpenRouter client initialization
   - Register with LLM manager

4. Add configuration in `/backend/internal/config/config.go`:
   - `OPENROUTER_API_KEY` environment variable
   - Optional `OPENROUTER_SITE_URL` for attribution

## Acceptance Criteria

- [ ] Provider implements full `Provider` interface
- [ ] Streaming responses work correctly
- [ ] Tool calling works for supported models
- [ ] Vision support for multimodal models
- [ ] API key validation works
- [ ] Can list all available models dynamically
- [ ] Popular models are pre-defined with correct capabilities
- [ ] No security vulnerabilities

## Files to Create/Modify

- `backend/internal/llm/openrouter/client.go` - Create OpenRouter client
- `backend/internal/llm/openrouter/models.go` - Create model definitions
- `backend/cmd/server/main.go` - Register provider
- `backend/internal/config/config.go` - Add config (if needed)

## Integration Points

- **Provides**: OpenRouter provider for LLM Manager
- **Consumes**: LLM Provider interface
- **Conflicts**: None - independent provider implementation

## API Reference

```
Base URL: https://openrouter.ai/api/v1
Auth: Authorization: Bearer <API_KEY>
Format: OpenAI-compatible (chat completions)
Headers: HTTP-Referer (optional), X-Title (optional)
```
