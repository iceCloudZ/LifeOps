# Backend Spike Comparison

## Summary

Two equivalent spikes were implemented:

- Quarkus + Java/GraalVM route
- Go standard-library route

Both implemented:

- `POST /api/inbox/webhook`
- In-memory inbox item creation
- AI draft JSON parser / output guard
- Automated tests

## Measurement Table

| Metric | Quarkus JVM | Quarkus Native | Go |
|---|---:|---:|---:|
| Test command | `mvnw test` | native runtime verified on 101 | `go test ./...` |
| Startup to first successful webhook request | 3.1s | 57ms on 101 | 25ms on 101 |
| Framework startup log | 4.2s test startup | 0.017s | not applicable |
| Working Set / RSS | 168.6 MB | 48.4 MB | 6.2 MB |
| Private Memory | 228.8 MB | not measured | not measured |
| App artifact size | 18.1 MB `quarkus-app` | 60 MB executable | 8.4 MB executable |
| linux/amd64 build | not verified | verified on 101 | 8.4 MB |
| linux/arm64 build | not verified | not verified | 7.8 MB |
| Native build | not applicable | 275s full Maven native build on 101, peak RSS 3.03 GB | not needed |

## Quarkus Interpretation

Quarkus is viable as a lighter Java route:

- It is clearly more self-hosted-friendly than a typical Spring Boot JVM app.
- It preserves Java ecosystem access.
- It supports GraalVM Native Image in principle.
- Native runtime is fast and much lighter than JVM mode.

The local Windows machine could not complete native build because Windows requires Visual Studio Build Tools:

```text
Failed to find 'vcvarsall.bat' in a Visual Studio installation.
```

The Linux `101` server completed native build successfully, but the first native run exposed a Jackson reflection issue. Adding `@RegisterForReflection` to DTO records fixed it.

This means Quarkus Native is viable, but it has real operational costs:

- Native build chain is heavier than Go.
- Build memory was about 3 GB for this tiny service.
- Native executable was about 60 MB.
- Native-specific reflection metadata must be maintained.

## Go Interpretation

Go fits the Homelab-first distribution story better:

- Low memory.
- Small binary.
- Fast startup.
- Easy cross-build for linux/amd64 and linux/arm64.
- No framework dependency for MVP APIs.
- Docker images can be tiny.

The MVP backend requirements are also simple enough for Go:

- REST API
- Webhook
- SQLite
- Scheduler
- MQTT
- OpenAI-compatible HTTP calls
- JSON output guard

## Recommendation

For the public LifeOps MVP:

> Prefer Go + Vue for the backend/frontend stack.

For fast internal exploration:

> Quarkus Native remains acceptable if Java reuse matters, but it should not be the default open-source pitch unless SQLite, MQTT, linux/arm64 native build, and reflection metadata maintenance are all verified.

## Next Step

If proceeding with Go:

1. Replace in-memory inbox with SQLite.
2. Add config loading for `LIFEOPS_WEBHOOK_TOKEN`.
3. Add `/api/inbox` list endpoint.
4. Add AI extractor interface with OpenAI-compatible client.
5. Add JSON Schema validation or strict parser rules.
6. Add Dockerfile with linux/amd64 and linux/arm64 build workflow.
