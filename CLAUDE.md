# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LifeOps is an open-source, self-hosted, local-first AI operating system for personal and family life. The first module is **FamilyOps**: a family AI inbox that turns messy household messages (e.g. WeChat) into structured drafts for tasks, events, shopping items, notes, and daily briefings.

Target users: NAS/Homelab users, Home Assistant users, privacy-sensitive tech families. Family data stays local; AI output is always user-confirmed before becoming real data.

## Build & Run Commands

All commands run from `apps/api/`:

```bash
cd apps/api
go test ./...                    # run all tests
go test -run TestWebhookInboxAcceptsMessage ./...   # run single test
go run .                         # start dev server on :8080
go build -trimpath -ldflags="-s -w" -o lifeops-api . # production build
```

Docker:

```bash
docker compose -f deploy/docker-compose.yml up -d
```

Cross-compile (as done in CI):

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o lifeops-api .
```

## Architecture

**Backend stack**: Go 1.26, zero external dependencies (stdlib only). Go was chosen over Quarkus Native for ~6 MB RSS vs ~48 MB.

**Entry point**: `apps/api/main.go` reads env vars and starts `Server` from `server.go`.

**Core flow**:
1. External systems POST text to `/api/inbox/webhook` (token-authenticated via `X-LifeOps-Token`)
2. `server.go` stores `InboxItem` in-memory (future: SQLite)
3. AI Extractor (not yet implemented) parses inbox content into structured `AIDraft` objects
4. `ai_draft_parser.go` validates AI output — JSON envelope with `drafts` array, each requiring `draft_type`, `title`, and `confidence`
5. User confirms drafts before they become real data

**Key design constraints**:
- All AI model output must go through the output guardrail (`ParseAIDrafts`) — never trust raw model text
- Draft types: `event`, `task`, `shopping_item`, `note`
- BYOK: users configure their own OpenAI-compatible endpoint (Ollama, DeepSeek, Qwen, etc.)
- SQLite write operations must be serialized (mutex or single-writer queue) for family-scale concurrency

**Environment variables**:
- `LIFEOPS_ADDR` — listen address (default `:8080`)
- `LIFEOPS_WEBHOOK_TOKEN` — webhook auth token (default `dev-token`)

## Repository Layout

- `apps/api/` — Go backend (MVP API service)
- `apps/web/` — Vue 3 + Vite + PWA frontend (planned)
- `internal/` — domain modules: family/inbox, drafts, tasks, events, briefing, ai/extractor, ai/guardrails (target structure, not yet migrated)
- `deploy/docker-compose.yml` — Docker deployment
- `docs/` — architecture, prompt strategy, backend spike comparisons
- `experiments/quarkus/` — Quarkus spike (legacy reference only)
- `familyops-mvp-architecture.md` — full MVP spec and data model

## Code Style

- Stdlib first, minimal external dependencies
- Organize code so Java engineers can read it: handler → service → store → model → config
- No complex frameworks or excessive Go abstractions
- AI prompt strategy is documented in `docs/prompt-strategy.md` — keep prompts and guardrails in sync
