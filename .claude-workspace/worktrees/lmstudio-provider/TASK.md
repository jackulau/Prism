---
id: lmstudio-provider
name: Add LM Studio Provider (Local)
wave: 1
priority: 2
dependencies: []
estimated_hours: 2
tags:
- backend
- llm-provider
- local
- open-source
---

## Objective

Implement LM Studio provider for local model inference via LM Studio's local server.

## Context

LM Studio is a popular desktop app for running local models:
- OpenAI-compatible API on localhost
- Easy model management UI
- Supports GGUF models
- Popular for local development

LM Studio exposes an OpenAI-compatible server on localhost:1234.

## Implementation

1. Create `/backend/internal/llm/lmstudio/client.go`:
   - Implement the `Provider` interface
   - Use OpenAI-compatible API format
   - Support streaming responses
   - Dynamically detect loaded model
   - No API key required (local)

2. Create `/backend/internal/llm/lmstudio/models.go`:
   - Fetch currently loaded model from LM Studio
   - Support model switching if multiple loaded

3. Add configuration:
   - `LMSTUDIO_HOST` environment variable (default: http://localhost:1234)

4. Register provider in `/backend/cmd/server/main.go`

## Acceptance Criteria

- [ ] Provider implements full `Provider` interface
- [ ] Streaming responses work correctly
- [ ] Detects currently loaded model
- [ ] Works without API key
- [ ] Configurable host/port
- [ ] Graceful handling when LM Studio not running
- [ ] No security vulnerabilities

## Files to Create/Modify

- `backend/internal/llm/lmstudio/client.go` - Create LM Studio client
- `backend/internal/llm/lmstudio/models.go` - Create model detection
- `backend/internal/config/config.go` - Add LMSTUDIO_HOST config
- `backend/cmd/server/main.go` - Register provider

## Integration Points

- **Provides**: LM Studio provider for LLM Manager
- **Consumes**: LLM Provider interface
- **Conflicts**: None - independent provider implementation

## API Reference

```
Base URL: http://localhost:1234/v1 (configurable)
Auth: None required (local)
Format: OpenAI-compatible (chat completions)
```
