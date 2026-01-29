---
id: fireworks-provider
name: Add Fireworks AI Provider
wave: 1
priority: 2
dependencies: []
estimated_hours: 3
tags:
- backend
- llm-provider
- open-source
- function-calling
---

## Objective

Implement Fireworks AI provider for fast inference with excellent function calling support.

## Context

Fireworks AI provides optimized inference for open source models:
- Llama 3.1/3.2/3.3 (all sizes)
- Mixtral models
- Qwen 2.5 models
- FireFunction: Specialized function calling model
- Yi models

Fireworks is known for excellent function calling reliability.

## Implementation

1. Create `/backend/internal/llm/fireworks/client.go`:
   - Implement the `Provider` interface
   - Use OpenAI-compatible API format
   - Support streaming responses
   - Implement tool calling (especially FireFunction)
   - Handle vision models

2. Create `/backend/internal/llm/fireworks/models.go`:
   - Define popular models with capabilities
   - Include FireFunction models for reliable tool use
   - Mark vision-capable models

3. Register provider in `/backend/cmd/server/main.go`

## Acceptance Criteria

- [ ] Provider implements full `Provider` interface
- [ ] Streaming responses work correctly
- [ ] Tool calling works reliably (especially FireFunction)
- [ ] Vision support for multimodal models
- [ ] API key validation works
- [ ] Popular models pre-defined
- [ ] No security vulnerabilities

## Files to Create/Modify

- `backend/internal/llm/fireworks/client.go` - Create Fireworks client
- `backend/internal/llm/fireworks/models.go` - Create model definitions
- `backend/cmd/server/main.go` - Register provider

## Integration Points

- **Provides**: Fireworks provider for LLM Manager
- **Consumes**: LLM Provider interface
- **Conflicts**: None - independent provider implementation

## API Reference

```
Base URL: https://api.fireworks.ai/inference/v1
Auth: Authorization: Bearer <API_KEY>
Format: OpenAI-compatible (chat completions)
```
