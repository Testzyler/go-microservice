# Project Agent Guide

## Persona
- you must read `architect.md` first to understand the high-level design.

## Project facts
- Go 1.25, Fiber HTTP.
- Internal gateway lives in `services/internal-gateway` (health, token issuance, auth middleware).
- Auth Connect service lives in `services/auth` (register/login/verify two-factor/get user, JWT issuance) backed by Postgres.
- Auth worker lives in `services/auth/cmd/worker` handling welcome/2FA email tasks via Asynq + Redis.
- Shared infra in `global/pkg` (config loader, logging, observability, token/auth, async/postgres scaffolding).
- Config via Viper + `.env` conventions, commands via Cobra.
- Observability: OpenTelemetry + Prometheus + Grafana + Tempo.
- Root CLI in `main.go` with subcommands: `serve-internal-gateway`, `serve-auth`, `serve-auth-worker`, `migrate`.

## Conventions
- Keep DDD boundaries when adding new services: domain is pure, application orchestrates, adapters do I/O.
- Gateway middleware lives under `services/internal-gateway/internal/middleware`.
- Shared utilities should go into `global/pkg` for reuse.
- If adding migrations later, use raw SQL `YYYYMMDDHHMM_name.up.sql` / `.down.sql`.

## Workflow
- Prefer `make` targets or `go run . <command>` for local runs.
- Add unit tests for application/domain changes.
- Update `README.md` when behavior or commands change.
