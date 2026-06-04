# Quarkus + GraalVM Spike

## Goal

Validate whether a lightweight Java route can support the LifeOps / FamilyOps MVP for self-hosted NAS and Home Assistant users.

This spike focuses on:

- Quarkus REST service.
- Webhook Inbox endpoint.
- AI draft output guard.
- JVM package size and runtime memory.
- Native Image build feasibility on the current machine.

## Environment

- OS: Windows
- Java: Oracle GraalVM 21.0.6
- Native Image: 21.0.6
- Maven: Maven Wrapper copied into this project
- Docker: installed, but Docker Desktop engine was not running during the spike
- Quarkus: 3.33.2

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

Commands:

```powershell
$env:JAVA_HOME='C:\Users\zzh58\.jdks\graalvm-jdk-21.0.6'
$env:Path="$env:JAVA_HOME\bin;$env:Path"
.\mvnw.cmd -s .\.mvn\central-settings.xml -q test
```

Result:

- `WebhookInboxResourceTest`: pass
- `AiDraftParserTest`: pass

## JVM Runtime Measurement

Command summary:

- Built JVM package with `.\mvnw.cmd -s .\.mvn\central-settings.xml -q package -DskipTests`
- Started `java -jar target\quarkus-app\quarkus-run.jar`
- Sent one webhook request.
- Read process memory.

Observed result:

| Metric | Value |
|---|---:|
| Startup to first successful webhook request | 3.1s |
| Working Set | 168.6 MB |
| Private Memory | 228.8 MB |
| `target/quarkus-app` size | 18.1 MB |

Interpretation:

- JVM Quarkus is lighter than a typical Spring Boot service.
- Memory is still higher than a Go service would likely be.
- For NAS / Homelab users, JVM mode is acceptable for some devices but not ideal as the default pitch.

## Native Image Result

### Windows Result

Command:

```powershell
.\mvnw.cmd -s .\.mvn\central-settings.xml package -Dnative -DskipTests
```

Result on Windows:

Native build reached GraalVM Native Image, then failed on Windows because Visual Studio Build Tools were missing:

```text
Error: Failed to find 'vcvarsall.bat' in a Visual Studio installation.
Please make sure that Visual Studio 2022 version 17.1.0 or later is installed on your system.
```

Interpretation:

- GraalVM Native Image is available locally.
- Windows native builds require Visual Studio Build Tools or an x64 Native Tools Command Prompt.
- Alternative: run native builds in Linux CI or via container build once Docker Desktop is running.
- Windows native build chain is more complex than Go.

### Linux 101 Result

The same project was built on the `101` Ubuntu server with Oracle GraalVM 21.0.11 and gcc.

Initial native runtime returned HTTP 500 because Jackson could not serialize `InboxItem` without native reflection metadata. Adding `@RegisterForReflection` to DTO records fixed it.

Native build result:

| Metric | Value |
|---|---:|
| Native image phase reported by Quarkus | 4m 13s |
| Full Maven native build after dependencies were present | 275s |
| Peak RSS during native-image build | 3.03 GB |
| Native executable size | 60 MB |

Runtime measurement on `101`:

| Metric | Value |
|---|---:|
| Quarkus startup log | 0.017s |
| Startup to first successful webhook request | 57ms |
| RSS | 48.4 MB |
| VSZ | 34.3 GB |
| Native executable size | 62,562,128 bytes |

Interpretation:

- Quarkus Native makes startup excellent and memory much lower than JVM mode.
- Native build requires substantial build memory and a more complex toolchain.
- DTO reflection hints are a real maintenance concern for native mode.
- Runtime still uses notably more memory and a much larger artifact than the equivalent Go spike.

## Preliminary Decision

Quarkus is viable for a Java route, but the spike confirms the trade-off:

- JVM mode is simple and works, but memory is around 170-230 MB for this tiny service.
- Native mode improves startup and memory dramatically, but build setup and reflection metadata add friction.
- For an open-source Homelab-first project, Go still has the cleaner default deployment story.
- For fastest personal development while preserving lower memory than Spring Boot, Quarkus is a reasonable Java option.

## Next Checks

Before choosing Java/Quarkus as the final backend:

1. Install Visual Studio Build Tools or use Linux CI to complete a native build.
2. Measure native executable size and memory.
3. Add SQLite and verify native compatibility.
4. Add MQTT publish and verify native compatibility.
5. Test linux/amd64 and linux/arm64 builds.
6. Compare against a minimal Go service with the same endpoints.
