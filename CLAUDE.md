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

- `apps/api-java/` — Java backend (Spring Boot 4.0.6, Java 25, SQLite, MyBatis-Plus, LangChain4j 1.16.0)
- `apps/api/` — Go backend (legacy, replaced by api-java)
- `apps/web/` — Vue 3 + Vite + PWA frontend
- `deploy/docker-compose.yml` — Docker deployment
- `docs/` — architecture, prompt strategy, specs, plans

## Build & Run (Java backend)

```bash
cd apps/api-java
JAVA_HOME="/c/Users/zzh58/.jdks/ms-25.0.3" ./mvnw compile     # compile
JAVA_HOME="/c/Users/zzh58/.jdks/ms-25.0.3" ./mvnw test         # run tests
JAVA_HOME="/c/Users/zzh58/.jdks/ms-25.0.3" ./mvnw spring-boot:run  # dev server
```

## Testing Rules

AI-generated code MUST be tested. This applies to all AI tools (Claude Code, Cursor, Copilot, etc.):

- **New public method → corresponding unit test**
- **New/modified Controller endpoint → integration test**
- **TDD preferred**: write failing test first, then implement
- **After any code change**: run `mvn test` to verify nothing breaks
- **Test directory**: `apps/api-java/src/test/java/com/lifeops/`
- **Frameworks**: JUnit 5, Mockito, Spring Boot Test (`@SpringBootTest`, `@WebMvcTest`)
- **Test naming**: `methodName_scenario_expectedResult` (e.g., `route_healthQuery_returnsHealthDomain`)

What to test:
- Unit: ToolDispatcher, LensRegistry, ButlerAgent routing logic, Service methods
- Integration: ChatController endpoints, ChatService two-phase flow
- Do NOT test trivial getters/setters or Lombok-generated code

## Code Style

- Java 25 idioms (records, sealed interfaces, text blocks, pattern matching)
- Organize: controller → service → mapper → entity → config
- No comments unless the WHY is non-obvious
- AI prompt strategy is documented in `docs/prompt-strategy.md`
