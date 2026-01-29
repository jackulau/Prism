---
id: mistral-provider
name: Add Mistral AI Provider (Direct API)
wave: 1
priority: 2
dependencies: []
estimated_hours: 3
tags:
- backend
- llm-provider
- open-source
---

## Objective

Implement Mistral AI provider for direct access to Mistral's models.

## Context

Mistral AI provides high-quality European AI models:
- Mistral Large 2 (flagship model)
- Mistral Small (efficient)
- Mistral NeMo (open weights)
- Codestral (code-specialized)
- Pixtral (vision model)
- Ministral (edge models)

Mistral uses an OpenAI-compatible API.

## Implementation

1. Create `/backend/internal/llm/mistral/client.go`:
   - Implement the `Provider` interface
   - Use OpenAI-compatible API format
   - Support streaming responses
   - Implement tool calling
   - Handle vision for Pixtral

2. Create `/backend/internal/llm/mistral/models.go`:
   - Define all available models
   - Include Codestral for coding tasks
   - Mark Pixtral as vision-capable

3. Register provider in `/backend/cmd/server/main.go`

## Acceptance Criteria

- [ ] Provider implements full `Provider` interface
- [ ] Streaming responses work correctly
- [ ] Tool calling works (native support)
- [ ] Vision support for Pixtral
- [ ] Codestral available for code tasks
- [ ] API key validation works
- [ ] No security vulnerabilities

## Files to Create/Modify

- `backend/internal/llm/mistral/client.go` - Create Mistral client
- `backend/internal/llm/mistral/models.go` - Create model definitions
- `backend/cmd/server/main.go` - Register provider

## Integration Points

- **Provides**: Mistral provider for LLM Manager
- **Consumes**: LLM Provider interface
- **Conflicts**: None - independent provider implementation

## API Reference

```
Base URL: https://api.mistral.ai/v1
Auth: Authorization: Bearer <API_KEY>
Format: OpenAI-compatible (chat completions)
```
