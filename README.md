# LifeOps

LifeOps is an open-source, self-hosted, local-first AI operating system for personal and family life.

The first module is **FamilyOps**: a family AI inbox that helps couples turn messy household messages into structured drafts for tasks, events, shopping items, notes, and daily briefings.

## Current Status

Early MVP / spike stage.

Implemented so far:

- Go backend spike under `apps/api`
- Webhook inbox endpoint
- AI draft JSON parser / output guard
- Backend spike comparison: Go vs Quarkus Native
- MVP architecture and prompt strategy docs

## Why LifeOps

LifeOps is designed for:

- Self-hosted NAS / Homelab users
- Home Assistant users
- Privacy-sensitive families
- BYOK and local model workflows

The core principle:

> Family data stays under family control. AI output is always reviewed before it becomes real household data.

## Backend Choice

The MVP backend is Go.

The decision is based on equivalent spikes:

- Go RSS on Linux: about 6 MB
- Quarkus Native RSS on Linux: about 48 MB
- Quarkus JVM RSS on Windows: about 169 MB

See:

- `docs/backend-spike-comparison.md`
- `docs/go-spike.md`
- `docs/quarkus-graalvm-spike.md`

## Run API Locally

```powershell
cd apps/api
go test ./...
go run .
```

Webhook example:

```bash
curl -X POST http://localhost:8080/api/inbox/webhook \
  -H "X-LifeOps-Token: dev-token" \
  -H "Content-Type: application/json" \
  -d '{"source":"wechat","sender":"partner","content":"周五孩子要带彩笔，别忘了交水费。"}'
```

## Docker

```bash
docker compose -f deploy/docker-compose.yml up -d
```

## Documentation

- `familyops-mvp-architecture.md`
- `docs/prompt-strategy.md`
- `docs/backend-spike-comparison.md`

## License

MIT
