# Go Microservice Template (auth + async scaffold)
First things first: please load `architect.md` to understand the intended structure and design approach.

This repo now holds a lean starting point: an **internal gateway** plus an **auth service skeleton** with shared async task contracts and multi-worker layout. It is meant to be extended with your own domain services while keeping logging, observability, and queueing patterns consistent.

## What’s included
- Fiber-based internal gateway (`services/internal-gateway`) with health, token issuance helper (`POST /api/v1/auth/token`), and auth middleware (`/api/v1/auth/me`).
- Auth service skeleton (`services/auth`) with per-service `TaskPublisher`, local task contracts under `services/auth/internal/adapters/queue`, Connect RPC handlers, and multi-worker/scheduler commands (`serve-auth-worker <mailer|auditor|cleanup>`, `serve-auth-scheduler`).
- Shared building blocks in `global/pkg` (config, logging, observability, token/auth, async client/server/publisher, postgres).
- Minimal CLI (`serve-internal-gateway`, `serve-auth`, `serve-auth-worker <worker>`, `serve-auth-scheduler`, optional `migrate` scaffold).
- Docker Compose for Postgres + Redis + gateway + auth + worker + Grafana/Prometheus/Tempo.
- Protos live in `proto`; generate Connect/gRPC clients/servers with `cd proto && buf generate` (output under `gen/proto`).

## Quick start
1) Copy envs (already duplicated):  
   `cp .env.example .env`
2) Run observability + gateway + auth + workers/scheduler:
   ```bash
   docker compose up -d
   # or locally without Docker
   go run . serve-internal-gateway
   go run . serve-auth
   go run . serve-auth-worker mailer
   go run . serve-auth-worker auditor
   go run . serve-auth-worker cleanup
   go run . serve-auth-scheduler
   ```
3) Issue a token from the gateway helper:
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/token \
     -H "Content-Type: application/json" \
     -d '{"user_id":"11111111-1111-1111-1111-111111111111","roles":["admin"],"permissions":["*"]}'
   ```
4) Auth is exposed via Connect (gRPC over HTTP) behind the gateway. Use gateway REST helpers:
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{"email":"user@example.com","password":"S3cretPass!"}'

   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"user@example.com","password":"S3cretPass!"}'
   ```

## Email (Postmark)
Mailer worker sends welcome/MFA emails via Postmark when enabled. Configure these envs:
```bash
AUTH_EMAIL_ENABLED=true
AUTH_EMAIL_FROM="Your Name <you@domain.com>"
AUTH_EMAIL_POSTMARK_SERVER_TOKEN=your-postmark-token
AUTH_EMAIL_POSTMARK_MESSAGE_STREAM=outbound
AUTH_EMAIL_POSTMARK_ENDPOINT=https://api.postmarkapp.com/email
```

## Structure
- `main.go` – CLI entry (gateway + auth + workers + scheduler + migrate).
- `services/internal-gateway` – Fiber server, middleware, and auth helpers.
- `services/auth` – auth skeleton (health endpoint for now) plus async TaskPublisher, worker handlers, and scheduler.
- `global/pkg` – shared infra (config, logging, observability, token/auth, async, postgres).
- `configs/` – Prometheus, Tempo, Grafana provisioning.
- `.env*` – environment defaults for local/Docker.

## Next steps
- Add your own services under `services/` and reuse `global/pkg` for config/logging/auth/observability.
- Expand the gateway to proxy or compose those services.
- Adjust Docker Compose or k8s manifests to match new services when you add them.
- Run Connect/proto generation once networked: `cd proto && buf generate` (config at `proto/buf.gen.yaml`).
- Run migrations for auth DB: `go run . migrate --path services/auth/migrations --database "$DATABASE_URL" --action up`.
