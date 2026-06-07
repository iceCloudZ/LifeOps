# AI Parsing Module Design

## Overview

The AI Parsing module converts unstructured family inbox text into structured drafts (tasks, events, shopping items, notes). It calls an OpenAI-compatible LLM endpoint with carefully engineered prompts and returns parsed results through the output guard layer.

## Interface

```go
type AIExtractor struct {
    baseURL    string
    model      string
    apiKey     string
    httpClient *http.Client
}

func NewAIExtractor(baseURL, model, apiKey string) *AIExtractor
func (e *AIExtractor) Extract(content string) (string, error)
```

`Extract` sends the inbox text to the LLM and returns the raw model output string. The output is then passed through the output guard (`ParseAIDrafts`) for validation.

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `LIFEOPS_LLM_BASE_URL` | OpenAI-compatible API base URL | (required for AI features) |
| `LIFEOPS_LLM_MODEL` | Model name | `qwen2.5:7b` |
| `LIFEOPS_LLM_API_KEY` | API key (empty for local models like Ollama) | (empty) |

## Supported Providers

Any OpenAI-compatible endpoint:

- **Ollama**: `http://localhost:11434/v1`
- **LM Studio**: `http://localhost:1234/v1`
- **DeepSeek**: `https://api.deepseek.com/v1`
- **OpenAI**: `https://api.openai.com/v1`
- **OpenRouter**: `https://openrouter.ai/api/v1`
- **Qwen (Tongyi)**: `https://dashscope.aliyuncs.com/compatible-mode/v1`

All use the same `/chat/completions` endpoint.

## Prompt Strategy

The system prompt and few-shot examples are documented in `docs/prompt-strategy.md`.

Key rules:
- System prompt defines the assistant as a family information organizer.
- Output must be a single JSON object with a top-level `drafts` array.
- Allowed `draft_type` values: `event`, `task`, `shopping_item`, `note`.
- Required fields per draft: `draft_type`, `title`, `confidence`.
- Optional fields: `description`, `due_at`, `assignee_hint`, `quantity`, `topic`.
- Model must use `null` for unknown values, not invent facts.

## API Call Flow

1. Build messages array: system prompt + few-shot examples + user message.
2. POST to `{baseURL}/chat/completions` with `model`, `messages`, and `temperature: 0`.
3. Extract `choices[0].message.content` from the response.
4. Return raw content string to caller.
5. Caller passes result through `ParseAIDrafts` (output guard).

## Error Handling

| Error | Behavior |
|-------|----------|
| Network failure | Return error, caller marks inbox item as `extraction_failed` |
| Non-200 API response | Return error with status code and body |
| Empty response content | Return error |
| Rate limit (429) | Return error, no automatic retry at this layer |

The retry-with-repair-prompt logic lives in the output guard layer, not here.

## Testing

- Unit tests use a mock HTTP server to simulate LLM responses.
- No real API calls in tests.
- Test cases: valid response, malformed JSON, empty response, network error, non-200 status.
