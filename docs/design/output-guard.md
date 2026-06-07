# Output Guard Design

## Overview

The output guard validates AI-generated output before it enters the system as a draft. It sits between the AI Extractor and draft storage, ensuring that only well-formed, schema-compliant data is accepted.

## Validation Pipeline

```
Raw model output (string)
  → JSON parse
  → Envelope validation (top-level "drafts" array)
  → Per-draft field validation
  → Return ValidatedDraft slice or structured error
```

## JSON Schema

The model must return exactly one JSON object:

```json
{
  "drafts": [
    {
      "draft_type": "task",
      "title": "...",
      "description": "...",
      "confidence": 0.91
    }
  ]
}
```

### Required fields per draft

| Field | Type | Constraint |
|-------|------|------------|
| `draft_type` | string | Must be one of: `event`, `task`, `shopping_item`, `note` |
| `title` | string | Non-empty after trim |
| `confidence` | float | Between 0.0 and 1.0 inclusive |

### Optional fields per draft

| Field | Type | Default |
|-------|------|---------|
| `description` | string | `""` |
| `due_at` | string/null | `null` |
| `assignee_hint` | string/null | `null` |
| `quantity` | string/null | `null` |
| `topic` | string/null | `null` |

## Validation Rules

1. **JSON parse**: Must be valid JSON. Failure → `invalid_json`.
2. **Envelope**: Must contain a top-level `drafts` array. Failure → `missing_drafts`.
3. **Non-empty**: `drafts` array must have at least one element. Failure → `missing_drafts`.
4. **Required fields**: Each draft must have `draft_type` and `title` (non-empty). Failure → `invalid_draft`.
5. **Type enum**: `draft_type` must be in the allowed set. Failure → `invalid_draft_type`.
6. **Confidence range**: `confidence` must be 0.0–1.0 if present. Failure → `invalid_confidence`.
7. **Sanitize**: Trim whitespace from strings. Default missing optional fields to empty/null.

## Repair Retry

If initial parsing fails with `invalid_json`, the guard performs one retry:

1. Send a repair prompt to the LLM:
   ```
   The previous response was not valid JSON. Convert it to exactly one valid JSON
   object with a top-level "drafts" array. Output JSON only. Do not include
   Markdown or explanations.
   ```
2. Re-parse the repaired response.
3. If still invalid, mark as `extraction_failed`.

This retry happens at the orchestration layer (caller of Extract + ParseAIDrafts), not inside ParseAIDrafts itself.

## Error Types

```go
type AIDraftParseResult struct {
    OK            bool
    Drafts        []AIDraft
    FailureReason string  // "invalid_json", "missing_drafts", "invalid_draft",
                          // "invalid_draft_type", "invalid_confidence"
}
```

## Sanitization

- Strip Markdown fences (```json ... ```) from model output before parsing.
- Trim leading/trailing whitespace.
- If the output starts with `{` and ends with `}` but has text before/after, extract the JSON object.

These sanitization steps help local small models that may add explanatory text around the JSON.
