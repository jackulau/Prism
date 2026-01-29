# LLM Provider Documentation

This guide covers all supported LLM providers in Prism, including setup instructions, model recommendations, and feature comparisons.

## Provider Overview

| Provider | API Key URL | Free Tier | Tool Calling | Vision | Max Context |
|----------|-------------|-----------|--------------|--------|-------------|
| OpenAI | [Get Key](https://platform.openai.com/api-keys) | No | Yes | Yes | 200K |
| Anthropic | [Get Key](https://console.anthropic.com/settings/keys) | No | Yes | Yes | 200K |
| Google AI | [Get Key](https://aistudio.google.com/app/apikey) | Yes | Yes | Yes | 1M |
| OpenRouter | [Get Key](https://openrouter.ai/keys) | No | Varies | Varies | Varies |
| Groq | [Get Key](https://console.groq.com/keys) | Yes | Yes | Some | 128K |
| DeepSeek | [Get Key](https://platform.deepseek.com/api_keys) | Yes | Yes | No | 64K |
| Ollama | [Download](https://ollama.ai) | Free | Some | Some | Varies |

---

## OpenAI

### Getting Started

1. Create an account at [platform.openai.com](https://platform.openai.com)
2. Go to [API Keys](https://platform.openai.com/api-keys)
3. Click "Create new secret key"
4. Copy the key and add it in Prism Settings

### Available Models

| Model | Description | Context | Best For |
|-------|-------------|---------|----------|
| o3 | Most powerful reasoning model | 200K | Complex reasoning, math |
| o4-mini | Fast, cost-efficient reasoning | 128K | Quick reasoning tasks |
| GPT-4.1 | Latest flagship model | 128K | General purpose |
| GPT-4.1 Mini | Fast and affordable | 128K | Simple tasks |
| GPT-4o | Legacy multimodal model | 128K | Backwards compatibility |

### Pricing

- GPT-4.1: ~$2.50/1M input, ~$10/1M output
- o3: ~$15/1M input, ~$60/1M output
- See [OpenAI Pricing](https://openai.com/pricing) for current rates

---

## Anthropic (Claude)

### Getting Started

1. Create an account at [console.anthropic.com](https://console.anthropic.com)
2. Go to [API Keys](https://console.anthropic.com/settings/keys)
3. Click "Create Key"
4. Copy the key and add it in Prism Settings

### Available Models

| Model | Description | Context | Best For |
|-------|-------------|---------|----------|
| Claude Opus 4.5 | Premium model, maximum intelligence | 200K | Complex analysis, research |
| Claude Sonnet 4.5 | Best balance of intelligence and speed | 200K | General purpose, coding |
| Claude Haiku 4.5 | Fastest with near-frontier intelligence | 200K | Quick tasks, high volume |

### Pricing

- Haiku: ~$0.25/1M input, ~$1.25/1M output
- Sonnet: ~$3/1M input, ~$15/1M output
- Opus: ~$15/1M input, ~$75/1M output
- See [Anthropic Pricing](https://www.anthropic.com/pricing) for current rates

### Recommendation

**Start with Claude Sonnet 4.5** for most tasks. It offers the best balance of capability and cost.

---

## Google AI (Gemini)

### Getting Started

1. Go to [Google AI Studio](https://aistudio.google.com/app/apikey)
2. Sign in with your Google account
3. Click "Create API key"
4. Copy the key and add it in Prism Settings

### Available Models

| Model | Description | Context | Best For |
|-------|-------------|---------|----------|
| Gemini 2.5 Pro | Best for complex reasoning | 1M | Research, long documents |
| Gemini 2.5 Flash | Fast and capable | 1M | General purpose |
| Gemini 2.0 Flash | Stable multimodal model | 1M | Production workloads |

### Pricing

- Free tier available with rate limits
- Gemini Flash: ~$0.075/1M input, ~$0.30/1M output
- Gemini Pro: ~$1.25/1M input, ~$5/1M output
- See [Google AI Pricing](https://ai.google.dev/pricing) for current rates

### Recommendation

**Gemini 2.5 Flash** is excellent for tasks involving long documents due to its 1M token context window.

---

## OpenRouter

### Getting Started

OpenRouter provides access to 200+ models from multiple providers through a single API.

1. Create an account at [openrouter.ai](https://openrouter.ai)
2. Go to [Keys](https://openrouter.ai/keys)
3. Click "Create Key"
4. Copy the key and add it in Prism Settings

### Popular Models

| Model | Description | Context | Best For |
|-------|-------------|---------|----------|
| Llama 3.3 70B | Meta's latest open model | 131K | General purpose |
| Llama 3.1 405B | Largest open model | 131K | Complex tasks |
| DeepSeek R1 | Strong reasoning | 64K | Math, coding |
| Mistral Large | Mistral's flagship | 128K | European languages |
| Mixtral 8x7B | High throughput MoE | 32K | Fast inference |
| Qwen 2.5 72B | Alibaba's powerful model | 131K | Multilingual |

### Pricing

- Pay-per-use based on each model's pricing
- Often cheaper than direct provider access
- See [OpenRouter Models](https://openrouter.ai/models) for pricing

### Recommendation

**OpenRouter is ideal if you want access to many models** without managing multiple API keys. Try Llama 3.3 70B for a free-tier-friendly option.

---

## Groq

### Getting Started

Groq offers extremely fast inference using custom LPU hardware.

1. Create an account at [console.groq.com](https://console.groq.com)
2. Go to [API Keys](https://console.groq.com/keys)
3. Click "Create API Key"
4. Copy the key and add it in Prism Settings

### Available Models

| Model | Description | Context | Vision |
|-------|-------------|---------|--------|
| Llama 3.3 70B Versatile | Best quality | 128K | No |
| Llama 3.2 90B Vision | Vision model | 128K | Yes |
| Llama 3.2 11B Vision | Efficient vision | 128K | Yes |
| Llama 3.1 8B Instant | Fastest text | 128K | No |
| Mixtral 8x7B | MoE model | 32K | No |
| Gemma 2 9B | Google's efficient model | 8K | No |

### Pricing

- Free tier: Limited requests per day
- Paid: ~$0.05-0.89/1M tokens depending on model
- See [Groq Pricing](https://console.groq.com/docs/pricing) for details

### Recommendation

**Groq is the fastest option for inference.** Use Llama 3.3 70B for quality or Llama 3.1 8B for maximum speed.

---

## DeepSeek

### Getting Started

DeepSeek offers high-quality models at very competitive prices.

1. Create an account at [platform.deepseek.com](https://platform.deepseek.com)
2. Go to [API Keys](https://platform.deepseek.com/api_keys)
3. Click "Create API Key"
4. Copy the key and add it in Prism Settings

### Available Models

| Model | Description | Context | Best For |
|-------|-------------|---------|----------|
| DeepSeek V3 | Flagship chat model | 64K | General purpose |
| DeepSeek R1 | Advanced reasoning | 64K | Math, complex reasoning |
| DeepSeek Coder | Code specialist | 64K | Programming tasks |

### Pricing

- Very competitive: ~$0.14/1M input, ~$0.28/1M output for V3
- R1: ~$0.55/1M input, ~$2.19/1M output
- See [DeepSeek Pricing](https://platform.deepseek.com/pricing) for current rates

### Recommendation

**DeepSeek offers the best value for money.** R1 is particularly strong for reasoning tasks at a fraction of the cost of comparable models.

---

## Ollama (Local)

### Getting Started

Ollama runs models locally on your machine - no API key needed.

1. Download Ollama from [ollama.ai](https://ollama.ai)
2. Install and start Ollama
3. Pull a model: `ollama pull llama3.2`
4. In Prism, select Ollama provider - models appear automatically

### Recommended Models

| Model | Size | RAM Needed | Best For |
|-------|------|------------|----------|
| llama3.2:3b | 3B | 4GB | Quick responses |
| llama3.2:latest | 11B | 8GB | Balanced |
| llama3.1:70b | 70B | 48GB | High quality |
| codellama | 7B-34B | 8-24GB | Code generation |
| qwen2.5-coder | 7B | 8GB | Coding |

### Setup Tips

- Ensure Ollama is running before using in Prism
- Default URL: `http://localhost:11434`
- Use `ollama list` to see installed models
- Tool calling requires Ollama 0.3.0+ and compatible models

### Recommendation

**Ollama is perfect for privacy-sensitive work** or when you want to avoid API costs. Start with `llama3.2` for a good balance of quality and speed.

---

## Model Recommendations by Use Case

### Coding & Programming
1. **Claude Sonnet 4.5** - Best overall for complex codebases
2. **DeepSeek Coder** - Great value, specialized for code
3. **GPT-4.1** - Strong general coding capabilities

### Research & Analysis
1. **Claude Opus 4.5** - Deepest analysis
2. **Gemini 2.5 Pro** - Best for long documents
3. **DeepSeek R1** - Strong reasoning at low cost

### Fast Responses
1. **Groq Llama 3.1 8B** - Fastest inference
2. **Claude Haiku 4.5** - Fast with high quality
3. **GPT-4.1 Mini** - Quick and capable

### Budget-Friendly
1. **DeepSeek V3** - Excellent value
2. **Groq** (free tier) - Fast and free
3. **Google Gemini Flash** - Good free tier
4. **Ollama** - Completely free (local)

### Privacy-First
1. **Ollama** - All data stays local
2. **Self-hosted models** via Ollama

---

## Troubleshooting

### "Invalid API Key" Error
- Verify the key is correctly copied (no extra spaces)
- Check the key hasn't expired or been revoked
- Ensure you have billing enabled (for paid providers)

### "Rate Limit" Error
- Wait and retry, or upgrade your plan
- Consider using a different provider temporarily

### Ollama Models Not Showing
- Ensure Ollama is running: `ollama list`
- Check the Ollama URL in settings (default: `http://localhost:11434`)
- Pull at least one model: `ollama pull llama3.2`

### Slow Responses
- Try a faster model (Haiku, GPT-4.1 Mini, or Groq)
- For Ollama, ensure sufficient RAM for your model size
