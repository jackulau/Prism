---
id: together-provider
name: Add Together AI Provider
wave: 1
priority: 1
dependencies: []
estimated_hours: 3
tags:
- backend
- llm-provider
- open-source
---

## Objective

Implement Together AI provider for serverless inference of open source models.

## Context

Together AI provides high-quality inference for many open source models:
- Meta Llama 3.1/3.2/3.3 (all sizes)
- Mistral/Mixtral models
- Qwen 2.5 models
- DeepSeek models
- CodeLlama
- Many specialized models

Together AI uses an OpenAI-compatible API.

## Implementation

1. Create `/backend/internal/llm/together/client.go`:
   - Implement the `Provider` interface
   - Use OpenAI-compatible API format
   - Support streaming responses
   - Implement tool calling for supported models
   - Handle vision models

2. Create `/backend/internal/llm/together/models.go`:
   - Define popular models with capabilities
   - Include accurate context windows
   - Fetch models dynamically from Together API

3. Register provider in `/backend/cmd/server/main.go`

## Acceptance Criteria

- [ ] Provider implements full `Provider` interface
- [ ] Streaming responses work correctly
- [ ] Tool calling works for supported models
- [ ] Vision support for multimodal models
- [ ] API key validation works
- [ ] Popular models pre-defined with correct capabilities
- [ ] No security vulnerabilities

## Files to Create/Modify

- `backend/internal/llm/together/client.go` - Create Together client
- `backend/internal/llm/together/models.go` - Create model definitions
- `backend/cmd/server/main.go` - Register provider

## Integration Points

- **Provides**: Together AI provider for LLM Manager
- **Consumes**: LLM Provider interface
- **Conflicts**: None - independent provider implementation

## API Reference

```
Base URL: https://api.together.xyz/v1
Auth: Authorization: Bearer <API_KEY>
Format: OpenAI-compatible (chat completions)
```
