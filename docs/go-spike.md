# Go Spike

## Goal

Implement the same minimal LifeOps / FamilyOps backend behavior as the Quarkus spike and compare runtime fit for NAS / Home Assistant / Homelab users.

This spike focuses on:

- Webhook Inbox endpoint.
- AI draft output guard.
- No third-party framework.
- Windows executable build.
- Linux amd64 and arm64 cross-builds.
- Startup, memory, and binary size.

## Environment

- OS: Windows
- Go: 1.26.4 portable zip under `.tools/go`
- Framework: Go standard library only
- Docker: not required

## Implemented Scope

### Webhook Inbox

Endpoint:

```text
POST /api/inbox/webhook
```

Headers:

```text
X-LifeOps-Token: dev-token
```

Request:

```json
{
  "source": "wechat",
  "sender": "partner",
  "content": "周五孩子要带彩笔，别忘了交水费。"
}
```

Behavior:

- Creates an in-memory inbox item.
- Returns `201`.
- Does not create tasks or events directly.

### AI Draft Parser

The parser accepts only structured JSON:

```json
{
  "drafts": [
    {
      "draft_type": "task",
      "title": "准备彩笔",
      "description": "周五孩子需要带彩笔",
      "confidence": 0.91
    }
  ]
}
```

Invalid non-JSON model output is rejected with `invalid_json`.

## Verification

Command:

```powershell
..\..\.tools\go\bin\go.exe test ./...
```

Result:

- `TestWebhookInboxAcceptsMessage`: pass
- `TestParseValidDraftJSON`: pass
- `TestRejectNonJSONModelOutput`: pass

## Runtime Measurement

Command summary:

- Built Windows executable with `go build`.
- Started `target/lifeops-go-spike.exe`.
- Sent one webhook request.
- Read process memory.

Observed result from second run:

| Metric | Value |
|---|---:|
| Startup to first successful webhook request | 0.26s |
| Working Set | 8.9 MB |
| Private Memory | 45.0 MB |
| Windows executable size | 8.4 MB |

First run after build showed 1.98s startup, likely affected by Windows process startup and initial file/cache behavior. The second run is more representative for an already-built executable.

## Cross-Build Result

Commands:

```powershell
$env:GOOS='linux'
$env:GOARCH='amd64'
..\..\.tools\go\bin\go.exe build -o ..\..\target\lifeops-go-spike-linux-amd64 .

$env:GOARCH='arm64'
..\..\.tools\go\bin\go.exe build -o ..\..\target\lifeops-go-spike-linux-arm64 .
```

Observed result:

| Target | Size |
|---|---:|
| linux/amd64 | 8.4 MB |
| linux/arm64 | 7.8 MB |

Interpretation:

- Go cross-builds cleanly without Docker.
- linux/amd64 and linux/arm64 are straightforward, which fits NAS and Homelab distribution.
- A scratch/distroless Docker image should be small and simple.

## Preliminary Decision

The Go spike is a better fit for a Homelab-first open-source backend:

- Much lower memory than JVM Quarkus mode.
- Smaller standalone executable.
- Faster startup.
- Cross-platform build story is simpler.
- No native-image toolchain or Visual Studio Build Tools issue.

The trade-off is that Java/Quarkus may still be faster for the primary developer if reusing existing Java agent code matters. For the public LifeOps MVP, Go has the cleaner deployment story.
