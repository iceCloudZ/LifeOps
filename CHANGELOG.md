# Changelog

All notable changes to LifeOps will be documented in this file.

## Unreleased

### Added

- Initial LifeOps repository.
- FamilyOps MVP architecture document.
- Prompt strategy document for FamilyOps AI Extractor.
- Go API service under `apps/api`.
- Webhook inbox endpoint for external message ingestion.
- AI draft JSON parser and output guard spike.
- Backend spike comparison for Go, Quarkus JVM, and Quarkus Native.
- Docker Compose and systemd deployment templates.
- GitHub Actions CI for Go tests and linux/amd64 plus linux/arm64 builds.
- GitHub Actions deployment workflow for the 101 server.

### Notes

- FamilyOps is the first MVP module of LifeOps.
- The deployed API currently binds to `127.0.0.1` on the 101 server and is not intended for public internet exposure.
- Quarkus remains in `experiments/quarkus` as a backend spike, while Go is the recommended MVP backend.
