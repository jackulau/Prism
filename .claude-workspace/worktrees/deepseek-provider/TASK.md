---
id: deepseek-provider
name: Add DeepSeek Provider (Direct API)
wave: 1
priority: 1
dependencies: []
estimated_hours: 3
tags:
- backend
- llm-provider
- open-source
- reasoning
---

## Objective

Implement DeepSeek provider for direct access to DeepSeek's powerful reasoning models.

## Context

DeepSeek provides state-of-the-art open source models:
- DeepSeek-V3: Flagship model with excellent coding and reasoning
- DeepSeek-R1: Advanced reasoning model (think before answering)
- DeepSeek-Coder: Specialized coding model

DeepSeek uses an OpenAI-compatible API with very competitive pricing.

## Implementation

1. Create `/backend/internal/llm/deepseek/client.go`:
   - Implement the `Provider` interface
   - Use OpenAI-compatible API format
   - Support streaming responses
   - Handle reasoning tokens (DeepSeek-R1 specific)
   - Implement tool calling

2. Create `/backend/internal/llm/deepseek/models.go`:
   - Define available models
   - DeepSeek-V3 (64k context)
   - DeepSeek-R1 series (reasoning models)
   - DeepSeek-Coder

3. Register provider in `/backend/cmd/server/main.go`

## Acceptance Criteria

- [ ] Provider implements full `Provider` interface
- [ ] Streaming responses work correctly
- [ ] Tool calling works
- [ ] Reasoning output handled for R1 models
- [ ] API key validation works
- [ ] Correct context windows (64k for V3)
- [ ] No security vulnerabilities

## Files to Create/Modify

- `backend/internal/llm/deepseek/client.go` - Create DeepSeek client
- `backend/internal/llm/deepseek/models.go` - Create model definitions
- `backend/cmd/server/main.go` - Register provider

## Integration Points

- **Provides**: DeepSeek provider for LLM Manager
- **Consumes**: LLM Provider interface
- **Conflicts**: None - independent provider implementation

## API Reference

```
Base URL: https://api.deepseek.com/v1
Auth: Authorization: Bearer <API_KEY>
Format: OpenAI-compatible (chat completions)
```
