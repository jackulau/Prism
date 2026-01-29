---
id: huggingface-provider
name: Add Hugging Face Inference API Provider
wave: 1
priority: 2
dependencies: []
estimated_hours: 4
tags:
- backend
- llm-provider
- open-source
- huggingface
---

## Objective

Implement Hugging Face Inference API provider for access to models hosted on Hugging Face Hub.

## Context

Hugging Face provides inference for thousands of open source models:
- Serverless Inference API (free tier available)
- Dedicated Inference Endpoints (for production)
- Access to any model on the Hub
- Popular models: Llama, Mistral, Falcon, StarCoder, etc.

The API format differs from OpenAI, requiring custom implementation.

## Implementation

1. Create `/backend/internal/llm/huggingface/client.go`:
   - Implement the `Provider` interface
   - Use Hugging Face Inference API format
   - Support streaming via Server-Sent Events
   - Handle the text-generation-inference format
   - Implement chat template handling

2. Create `/backend/internal/llm/huggingface/models.go`:
   - Define popular models available on HF Hub
   - Include models with chat templates
   - Support custom model endpoints

3. Register provider in `/backend/cmd/server/main.go`

## Acceptance Criteria

- [ ] Provider implements full `Provider` interface
- [ ] Streaming responses work correctly
- [ ] Chat templates applied correctly
- [ ] Tool calling for supported models
- [ ] API key validation works
- [ ] Can specify custom model IDs
- [ ] No security vulnerabilities

## Files to Create/Modify

- `backend/internal/llm/huggingface/client.go` - Create HF client
- `backend/internal/llm/huggingface/models.go` - Create model definitions
- `backend/cmd/server/main.go` - Register provider

## Integration Points

- **Provides**: Hugging Face provider for LLM Manager
- **Consumes**: LLM Provider interface
- **Conflicts**: None - independent provider implementation

## API Reference

```
Base URL: https://api-inference.huggingface.co/models/{model_id}
           https://api-inference.huggingface.co/v1/chat/completions (new format)
Auth: Authorization: Bearer <HF_TOKEN>
Format: Hugging Face text-generation format or OpenAI-compatible
```
