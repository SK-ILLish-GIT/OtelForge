# OtelForge

Internal platform to deploy and operate OpenTelemetry Collector configurations across SSH-reachable Linux nodes.

Register instances, launch **events** against one or many targets, and watch **live job output** in the UI. Admins audit activity across users.

**Full design, flows, data model, API, security, and roadmap:** see [ARCHITECTURE.md](ARCHITECTURE.md).

## Quick start

```bash
cp .env.example .env
# Set unique JWT_SECRET and ENCRYPTION_KEY before non-local use.

docker compose up -d --build
```

Jobs stay **QUEUED** if the worker is not running:

```bash
docker compose up -d worker
```

Seed an admin (first run):

```bash
export DATABASE_URL=postgres://otel:otel@localhost:5432/otelforge?sslmode=disable
export ENCRYPTION_KEY=0123456789abcdef0123456789abcdef   # must match .env
go run ./cmd/seed --email admin@internal.local --password changeme --role admin
```

| Service | URL |
|---------|-----|
| Web UI | http://localhost:5173 |
| API | http://localhost:8080/health |
| RabbitMQ | http://localhost:15672 (`otel` / `otel`) |

Default login: `admin@internal.local` / `changeme` — change before shared use.

## Local development

```bash
docker compose up postgres rabbitmq -d
go run ./cmd/seed --email admin@internal.local --password changeme --role admin

go run ./api/cmd/server          # terminal 1
go run ./worker/cmd/worker       # terminal 2
cd web && npm install && npm run dev   # terminal 3
```

## Configuration

Copy `.env.example` → `.env`. Required: `JWT_SECRET`, `ENCRYPTION_KEY` (32 hex chars), `DATABASE_URL`, `RABBITMQ_URL`, `CORS_ORIGIN`.

## Project layout

```
api/           Go REST API (Fiber)
worker/        RabbitMQ consumer + SSH runner
web/           React + Vite UI
internal/      Shared packages (db, auth, queue, ssh, crypto)
scripts/bash/  Platform-owned remote task scripts
cmd/seed/      Create users from CLI
ARCHITECTURE.md
```
