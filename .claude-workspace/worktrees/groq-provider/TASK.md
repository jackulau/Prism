---
id: groq-provider
name: Add Groq Provider for Fast Inference
wave: 1
priority: 1
dependencies: []
estimated_hours: 3
tags:
- backend
- llm-provider
- open-source
- fast-inference
---

## Objective

Implement Groq provider for extremely fast inference of open source models using Groq's LPU hardware.

## Context

Groq provides the fastest inference speeds for open source models:
- Llama 3.1 70B, 8B
- Llama 3.2 (including vision models)
- Llama 3.3 70B
- Mixtral 8x7B
- Gemma 2 9B

Groq uses an OpenAI-compatible API, making integration straightforward.

## Implementation

1. Create `/backend/internal/llm/groq/client.go`:
   - Implement the `Provider` interface
   - Use OpenAI-compatible API format
   - Support streaming responses
   - Implement tool calling
   - Handle vision models (Llama 3.2 vision)

2. Create `/backend/internal/llm/groq/models.go`:
   - Define available models with capabilities
   - Include accurate context windows (8k-128k depending on model)
   - Mark tool support and vision support

3. Register provider in `/backend/cmd/server/main.go`

## Acceptance Criteria

- [ ] Provider implements full `Provider` interface
- [ ] Streaming responses work at high speed
- [ ] Tool calling works for supported models
- [ ] Vision support for Llama 3.2 vision
- [ ] API key validation works
- [ ] Correct context windows for each model
- [ ] No security vulnerabilities

## Files to Create/Modify

- `backend/internal/llm/groq/client.go` - Create Groq client
- `backend/internal/llm/groq/models.go` - Create model definitions
- `backend/cmd/server/main.go` - Register provider

## Integration Points

- **Provides**: Groq provider for LLM Manager
- **Consumes**: LLM Provider interface
- **Conflicts**: None - independent provider implementation

## API Reference

```
Base URL: https://api.groq.com/openai/v1
Auth: Authorization: Bearer <API_KEY>
Format: OpenAI-compatible (chat completions)
```
