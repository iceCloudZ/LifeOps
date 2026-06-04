# Prompt Strategy for FamilyOps AI Extractor

## Goal

FamilyOps AI Extractor turns messy family inbox text into structured drafts:

- `event`
- `task`
- `shopping_item`
- `note`

The extractor must work with both strong hosted models and smaller local models such as 7B Qwen-family models. The product must never trust raw model text directly. Every model response goes through output guardrails and user confirmation.

## Non-Goals

The extractor does not:

- Create confirmed tasks or events directly.
- Give medical, investment, insurance sales, legal, psychological, or relationship judgment advice.
- Decide which spouse is right or wrong.
- Store or expose full prompts in audit logs.

## Output Contract

The model must return one JSON object and nothing else:

```json
{
  "drafts": [
    {
      "draft_type": "task",
      "title": "准备彩笔",
      "description": "周五孩子需要带彩笔",
      "due_at": null,
      "assignee_hint": null,
      "confidence": 0.91
    }
  ]
}
```

Allowed `draft_type` values:

- `event`
- `task`
- `shopping_item`
- `note`

Required fields:

- `draft_type`
- `title`
- `confidence`

Optional fields:

- `description`
- `due_at`
- `assignee_hint`
- `quantity`
- `topic`

## System Prompt Principles

The system prompt should:

- State that the assistant is a family information organizer, not a decision maker.
- Require strict JSON output.
- Define the allowed draft types.
- Tell the model to preserve uncertainty instead of inventing facts.
- Tell the model not to infer exact dates when the date is ambiguous.
- Tell the model to use `null` for unknown values.
- Tell the model not to output Markdown.
- Tell the model not to explain the JSON.

## Few-Shot Examples

### Example 1: Task + Task

Input:

```text
周五孩子要带彩笔，别忘了交水费。
```

Output:

```json
{
  "drafts": [
    {
      "draft_type": "task",
      "title": "准备彩笔",
      "description": "周五孩子需要带彩笔",
      "due_at": null,
      "assignee_hint": null,
      "confidence": 0.9
    },
    {
      "draft_type": "task",
      "title": "交水费",
      "description": "提醒家庭处理水费缴纳",
      "due_at": null,
      "assignee_hint": null,
      "confidence": 0.86
    }
  ]
}
```

### Example 2: Event

Input:

```text
下周二 18:30 家长会，我可能来不及，你看能不能去。
```

Output:

```json
{
  "drafts": [
    {
      "draft_type": "event",
      "title": "家长会",
      "description": "下周二 18:30 有家长会，需要确认谁参加",
      "due_at": null,
      "assignee_hint": null,
      "confidence": 0.82
    },
    {
      "draft_type": "task",
      "title": "确认家长会参加人",
      "description": "需要确认伴侣是否能参加家长会",
      "due_at": null,
      "assignee_hint": "partner",
      "confidence": 0.78
    }
  ]
}
```

### Example 3: Shopping Items

Input:

```text
周日买牛奶、鸡蛋、纸巾，晚上少油少辣。
```

Output:

```json
{
  "drafts": [
    {
      "draft_type": "shopping_item",
      "title": "牛奶",
      "description": null,
      "quantity": null,
      "confidence": 0.95
    },
    {
      "draft_type": "shopping_item",
      "title": "鸡蛋",
      "description": null,
      "quantity": null,
      "confidence": 0.95
    },
    {
      "draft_type": "shopping_item",
      "title": "纸巾",
      "description": null,
      "quantity": null,
      "confidence": 0.95
    },
    {
      "draft_type": "note",
      "title": "晚餐偏好",
      "description": "晚上少油少辣",
      "topic": "food",
      "confidence": 0.88
    }
  ]
}
```

## Local Small Model Tactics

Small local models may add prose, Markdown, or malformed JSON. Use these tactics:

- Prefer OpenAI-compatible JSON mode when the provider supports it.
- Keep the schema small.
- Use few-shot examples with short outputs.
- Put the JSON-only instruction near the end of the prompt.
- Consider pre-filling the assistant response with `{` when the model adapter supports it.
- Retry once with a shorter repair prompt when parsing fails.

Repair prompt idea:

```text
The previous response was not valid JSON. Convert it to exactly one valid JSON object with a top-level "drafts" array. Output JSON only. Do not include Markdown or explanations.
```

## Guardrail Flow

1. Build prompt.
2. Call configured model.
3. Parse response as JSON.
4. Validate top-level `drafts` array.
5. Validate each draft has `draft_type`, `title`, and `confidence`.
6. Reject unknown draft types.
7. If parsing fails, retry once with repair prompt.
8. If retry fails, mark inbox item as `extraction_failed`.
9. User can manually create drafts from the inbox item.

## Audit Logging

Audit logs should record:

- provider
- model
- purpose
- input reference id
- success or failure
- failure reason
- token usage if available

Audit logs should not store full sensitive prompts or raw model outputs by default.

## Open Questions

- Should `due_at` be accepted as natural language during MVP, or only ISO timestamp / null?
- Should the extractor split shopping items into separate drafts or preserve one combined shopping list draft?
- Should family-specific member names be included in the prompt, or should assignee detection be postponed?
- Should low-confidence drafts be shown separately from normal drafts?
