---
id: openai-compatible-provider
name: Add Generic OpenAI-Compatible Provider
wave: 1
priority: 1
dependencies: []
estimated_hours: 4
tags:
- backend
- llm-provider
- local
- open-source
- generic
---

## Objective

Implement a generic OpenAI-compatible provider that users can configure for any OpenAI-compatible endpoint.

## Context

Many services provide OpenAI-compatible APIs:
- vLLM servers
- llama.cpp server
- LocalAI
- text-generation-webui
- Jan
- Self-hosted inference servers
- Custom deployments

A generic provider allows users to add any compatible endpoint without code changes.

## Implementation

1. Create `/backend/internal/llm/openaicompat/client.go`:
   - Implement the `Provider` interface
   - Configurable base URL
   - Optional API key
   - Support streaming responses
   - Tool calling (if endpoint supports it)
   - User-configurable model list

2. Create `/backend/internal/llm/openaicompat/models.go`:
   - Fetch models from /v1/models endpoint if available
   - Allow manual model configuration
   - Support custom model names

3. Create database table for custom endpoints:
   - Store base URL, API key, name, models
   - Per-user custom endpoints

4. Create API endpoints:
   - `POST /api/v1/providers/custom` - Add custom endpoint
   - `GET /api/v1/providers/custom` - List custom endpoints
   - `DELETE /api/v1/providers/custom/:id` - Remove endpoint

5. Update frontend to allow adding custom providers

## Acceptance Criteria

- [ ] Can configure any OpenAI-compatible endpoint
- [ ] Streaming responses work
- [ ] Tool calling works (when supported)
- [ ] Can add multiple custom endpoints
- [ ] API keys encrypted in database
- [ ] Can fetch models from endpoint
- [ ] Frontend UI for configuration
- [ ] No security vulnerabilities

## Files to Create/Modify

- `backend/internal/llm/openaicompat/client.go` - Generic client
- `backend/internal/llm/openaicompat/models.go` - Model handling
- `backend/internal/database/repository/custom_provider.go` - DB repository
- `backend/internal/api/handlers/custom_provider_handler.go` - API handlers
- `backend/internal/api/routes/router.go` - Add routes
- `frontend/src/components/CustomProviderSettings.tsx` - UI component

## Integration Points

- **Provides**: Flexible provider for any OpenAI-compatible endpoint
- **Consumes**: LLM Provider interface, Database
- **Conflicts**: None - independent implementation

## Database Schema

```sql
CREATE TABLE custom_providers (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    base_url TEXT NOT NULL,
    encrypted_api_key TEXT,
    models TEXT, -- JSON array
    supports_tools BOOLEAN DEFAULT FALSE,
    supports_vision BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```
