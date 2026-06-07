# Webhook Inbox Design

## Overview

The Webhook Inbox is the primary entry point for external messages into FamilyOps. External adapters (WeChat bridges, Telegram bots, Home Assistant automations, custom scripts) POST text messages to a single endpoint. Messages are stored as `inbox_item` records and later processed by the AI Extractor.

## Endpoint

```
POST /api/inbox/webhook
```

## Authentication

All webhook requests must include a shared secret token via header:

```
X-LifeOps-Token: <token>
```

- The token is set via the `LIFEOPS_WEBHOOK_TOKEN` environment variable.
- Default value for development: `dev-token`.
- Requests without a valid token receive `401 Unauthorized`.
- MVP uses a single shared token. Per-source tokens and HMAC signature verification are deferred to a later phase.

## Request Body

```json
{
  "source": "wechat",
  "sender": "partner",
  "content": "周五孩子要带彩笔，别忘了交水费。"
}
```

### Field Definitions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source` | string | yes | Identifier for the message channel (e.g., `wechat`, `telegram`, `homeassistant`, `manual`) |
| `sender` | string | no | Identifier for the sender within the source context (e.g., `partner`, `self`, user ID) |
| `content` | string | yes | The raw text content of the message |

### Source Types

MVP defines these source identifiers:

| Source | Description |
|--------|-------------|
| `wechat` | Messages bridged from WeChat via tools like WeChatFerry |
| `telegram` | Messages from a Telegram bot |
| `homeassistant` | Messages pushed from Home Assistant automations |
| `manual` | Messages entered directly by the user through the web UI |

Additional sources can be added without code changes — the `source` field is a free-form string. MVP does not validate against a fixed list, but the web UI and docs should guide users toward these standard values.

### Content Constraints

- `content` must be a non-empty string after trimming whitespace.
- Maximum length: 4096 characters (configurable). Longer content is truncated with a warning.
- MVP only supports plain text. Images, voice, and structured attachments are out of scope.

## Response

### Success (201 Created)

```json
{
  "id": "a1b2c3d4e5f6...",
  "source": "wechat",
  "sender": "partner",
  "content": "周五孩子要带彩笔，别忘了交水费。",
  "status": "new",
  "created_at": "2026-06-08T10:30:00Z"
}
```

### Error Responses

| Status | Condition | Body |
|--------|-----------|------|
| 400 Bad Request | Invalid JSON, missing required fields, or empty content | `{"error": "invalid_request", "message": "..."}` |
| 401 Unauthorized | Missing or invalid `X-LifeOps-Token` | (empty body) |
| 405 Method Not Allowed | Non-POST request | (empty body) |

### Error Response Format

```json
{
  "error": "invalid_request",
  "message": "content is required"
}
```

`error` is a machine-readable code. `message` is a human-readable description.

Error codes:

- `invalid_request` — malformed JSON or missing required fields
- `empty_content` — `content` is empty or whitespace-only

## Storage

### MVP (In-Memory)

The current spike stores `inbox_item` records in a Go slice protected by `sync.Mutex`. This is sufficient for single-instance development but does not survive restarts.

### SQLite (Phase 1 Target)

Inbox items should be persisted to SQLite with this schema:

```sql
CREATE TABLE inbox_items (
    id          TEXT PRIMARY KEY,
    family_id   TEXT NOT NULL,
    source      TEXT NOT NULL,
    sender      TEXT,
    content     TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'new',
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);
```

Status transitions: `new` → `extracted` → `confirmed` / `archived`.

SQLite writes must be serialized (mutex or single-writer queue) to avoid `database is locked` errors under concurrent family use.

## Idempotency

MVP does not require duplicate detection. The same message can create multiple inbox items. Idempotency keys and deduplication are deferred to a later phase.

## Rate Limiting

MVP does not implement rate limiting. A single family with two adults and a few automation sources is unlikely to overwhelm the endpoint. Rate limiting can be added later if needed, either via middleware or a reverse proxy (Caddy/Nginx).

## Security Considerations

- The webhook endpoint must only be accessible within the trusted home network.
- Do not expose the endpoint directly to the public internet. Use Tailscale, WireGuard, or a reverse proxy with HTTPS if remote access is needed.
- The shared token provides basic authentication but is not a substitute for network-level security.
- Future enhancements: per-source tokens, HMAC signature verification, IP allowlists.
