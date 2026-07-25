# OtelForge

Internal platform to deploy and operate OpenTelemetry Collector configurations across SSH-reachable Linux nodes.

Operators register instances, launch **events** (batch jobs) against one or many targets, and watch live execution output. Admins can audit activity across users.

## Features

- Register SSH targets (single form or CSV bulk import)
- Launch platform-owned tasks: deploy, validate, install agent, restart, rollback, logs, SSH test, and more
- Async job execution via RabbitMQ with per-node results, checks, and stdout/stderr
- Live event console with job timeline and polling
- User and admin roles with JWT authentication
- Encrypted SSH credentials at rest (AES-256-GCM)
- SSH host key pinning (TOFU) on first connect

## Architecture

```
React UI  →  Go API  →  PostgreSQL
                ↓
            RabbitMQ  →  Go Worker  →  SSH + bash scripts on target nodes
```

| Component | Path | Role |
|-----------|------|------|
| API | `api/cmd/server` | REST API, auth, enqueue jobs |
| Worker | `worker/cmd/worker` | Consume queue, SSH to nodes, run scripts |
| Web | `web/` | React + Vite operator UI |
| Scripts | `scripts/bash/` | Platform-owned remote task scripts |
| Shared | `internal/` | DB, queue, crypto, SSH, models |

## Prerequisites

- Docker and Docker Compose
- Go 1.25+ (local dev without Docker)
- Node.js 20+ (local web dev)

## Quick start (Docker)

```bash
cp .env.example .env
# Set unique JWT_SECRET and ENCRYPTION_KEY before any non-local use.

docker compose up -d --build
```

**Important:** Jobs remain `QUEUED` if the worker is not running. Always run the full stack detached (`-d`), or start the worker explicitly:

```bash
docker compose up -d worker
```

Seed an admin user (first run only):

```bash
export DATABASE_URL=postgres://otel:otel@localhost:5432/otelforge?sslmode=disable
export ENCRYPTION_KEY=0123456789abcdef0123456789abcdef   # must match .env
go run ./cmd/seed --email admin@internal.local --password changeme --role admin
```

Open:

| Service | URL |
|---------|-----|
| Web UI | http://localhost:5173 |
| API health | http://localhost:8080/health |
| RabbitMQ management | http://localhost:15672 (user/pass: `otel` / `otel`) |

Default seeded login: `admin@internal.local` / `changeme` — change before shared use.

## Local development (without full Docker stack)

```bash
cp .env.example .env

# Infrastructure only
docker compose up postgres rabbitmq -d

# Seed admin
go run ./cmd/seed --email admin@internal.local --password changeme --role admin

# Terminals
go run ./api/cmd/server
go run ./worker/cmd/worker
cd web && npm install && npm run dev
```

## Configuration

Copy `.env.example` to `.env`. Required variables:

| Variable | Purpose |
|----------|---------|
| `JWT_SECRET` | Signs API session tokens |
| `ENCRYPTION_KEY` | 32 hex chars (16 bytes) for SSH secret encryption |
| `DATABASE_URL` | Postgres connection string |
| `RABBITMQ_URL` | RabbitMQ AMQP URL |
| `CORS_ORIGIN` | Allowed browser origin for API |

Docker Compose loads `.env` for the API and worker. Postgres and RabbitMQ defaults in compose are for **local development only**.

## Project layout

```
api/           HTTP API (Fiber)
worker/        Queue consumer + SSH runner
web/           React operator UI
internal/      Shared Go packages
scripts/bash/  Remote task scripts
cmd/seed/      CLI to create users
docs/          Design specs and plans
```

## Security notes (production)

Before exposing beyond a trusted network:

- Generate unique `JWT_SECRET` and `ENCRYPTION_KEY`
- Do not publish Postgres or RabbitMQ ports publicly
- Terminate TLS at a reverse proxy
- Review SSH host pinning and instance host allowlisting for your environment
- Rotate seeded credentials; use `cmd/seed` to provision operators

## Docs

- [v1 Design Spec](docs/superpowers/specs/2026-07-26-otelforge-v1-design.md)
- [Implementation Plan](docs/superpowers/plans/2026-07-26-otelforge-v1-implementation-plan.md)
- [Task Tracker](docs/superpowers/plans/2026-07-26-otelforge-v1-task-tracker.md)
- [Architecture (legacy import)](docs/otel-deploy/ARCHITECTURE.md)

## License

Private / internal use unless otherwise specified by the repository owner.
