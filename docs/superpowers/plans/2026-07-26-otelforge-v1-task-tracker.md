# OtelForge v1 — Task Tracker

> **Spec:** [2026-07-26-otelforge-v1-design.md](../specs/2026-07-26-otelforge-v1-design.md)  
> **Last updated:** 2026-07-26  
> **How to use:** Check `- [x]` when done. Update the progress table after each phase.

## Progress summary

| Phase | Done | Total | Status |
|-------|------|-------|--------|
| 0 — Planning & docs | 6 | 6 | ✅ Complete |
| 1 — Repo & infrastructure | 2 | 10 | 🔴 Not started |
| 2 — Shared core (`internal/`) | 0 | 12 | 🔴 Not started |
| 3 — Go API | 0 | 18 | 🔴 Not started |
| 4 — Go worker | 0 | 16 | 🔴 Not started |
| 5 — Bash task scripts | 0 | 8 | 🔴 Not started |
| 6 — React UI | 0 | 14 | 🔴 Not started |
| 7 — Integration & hardening | 0 | 10 | 🔴 Not started |
| **Total** | **8** | **94** | **9%** |

---

## Phase 0 — Planning & docs

- [x] Original architecture doc imported
- [x] v1 design spec written (`2026-07-26-otelforge-v1-design.md`)
- [x] Architecture doc aligned to Postgres + v1 scope
- [x] BMAD Method installed
- [x] Task tracker created (this doc)
- [x] Implementation plan written ([2026-07-26-otelforge-v1-implementation-plan.md](./2026-07-26-otelforge-v1-implementation-plan.md))

---

## Phase 1 — Repo & infrastructure

### Monorepo scaffold

- [x] `docker-compose.yml` (Postgres + RabbitMQ + api/worker/web stubs)
- [x] `.env.example`
- [x] `.gitignore`
- [ ] Root `go.mod` (monorepo module path)
- [ ] `api/cmd/server/main.go` entrypoint
- [ ] `worker/cmd/worker/main.go` entrypoint
- [ ] `api/Dockerfile`
- [ ] `worker/Dockerfile`
- [ ] `web/` Vite + React + TS scaffold
- [ ] `web/Dockerfile`
- [ ] Root `README.md` (local dev instructions)

### Docker Compose services

- [ ] Postgres migrations run on API startup (or init container)
- [ ] RabbitMQ exchange + queues declared on startup
- [ ] `docker compose up` brings full stack up healthy
- [ ] Seed script for admin user (`cmd/seed` or Makefile target)

---

## Phase 2 — Shared core (`internal/`)

### Config & models

- [ ] `internal/config` — env loading (`DATABASE_URL`, `RABBITMQ_URL`, secrets)
- [ ] `internal/models` — User, Instance, Event, Job, JobCheck, enums
- [ ] Task type enum (8 predefined tasks)

### Database

- [ ] `internal/db/migrations/001_initial.sql` — users, instances, events, jobs, job_checks
- [ ] `internal/db` — connection pool, query helpers
- [ ] Indexes for admin filters (launcher_email, status, created_at, task_type)

### Crypto

- [ ] `internal/crypto/bcrypt` — user password hashing
- [ ] `internal/crypto/aes` — SSH password encrypt/decrypt (AES-256-GCM)
- [ ] Encryption key from env; never log plaintext

### Queue

- [ ] `internal/queue` — RabbitMQ connect, declare topology
- [ ] Exchange `otel.deploy`, queue `deploy.jobs`, DLQ `deploy.jobs.dlq`
- [ ] Publish job message `{ jobId, eventId, instanceId, taskType }`
- [ ] Consumer with prefetch=1, ACK after DB update, NACK + retry (max 3)

---

## Phase 3 — Go API

### Server foundation

- [x] HTTP router (Go Fiber) + `/api/v1` prefix
- [ ] Request logging, panic recovery, CORS
- [ ] Health check `GET /health`

### Auth

- [ ] `POST /api/v1/auth/login` — email/password → JWT
- [ ] JWT middleware — validate token, attach user to context
- [ ] Role middleware — `RequireUser`, `RequireAdmin`
- [ ] No public register endpoint (seed-only users)

### Instances

- [ ] `GET /api/v1/instances` — list (own; admin: all)
- [ ] `POST /api/v1/instances` — create single node
- [ ] `POST /api/v1/instances/bulk` — CSV parse + bulk create
- [ ] `GET /api/v1/instances/:id` — detail (no password in response)
- [ ] `PUT /api/v1/instances/:id` — update
- [ ] `DELETE /api/v1/instances/:id` — delete
- [ ] Ownership check on all instance routes

### Events & jobs

- [ ] `POST /api/v1/events` — create event + jobs + enqueue
- [ ] Pre-run validation: YAML syntax (config tasks), instance ownership
- [ ] Auto-fill `launcherName`, `launcherEmail` from JWT user
- [ ] `GET /api/v1/events` — list (own; admin: all + query filters)
- [ ] Admin filters: launcher email, status, task type, date range
- [ ] `GET /api/v1/events/:id` — event + jobs + checks + logs
- [ ] `POST /api/v1/events/:id/rerun-failed` — re-queue failed jobs only
- [ ] `POST /api/v1/events/:id/clone` — copy config/task/nodes → new event
- [ ] `POST /api/v1/events/:id/jobs/:jobId/rollback` — per-node rollback event
- [ ] Event status rollup: COMPLETED | PARTIAL | FAILED from job counts

### Admin

- [ ] `GET /api/v1/admin/events` — all events with filters
- [ ] `GET /api/v1/admin/instances` — all instances with owner info

---

## Phase 4 — Go worker

### Worker foundation

- [ ] Consume from `deploy.jobs` queue
- [ ] Load job, event, instance from Postgres
- [ ] Decrypt SSH password in memory only
- [ ] Update job status lifecycle: QUEUED → RUNNING → VERIFIED | FAILED
- [ ] Write stdout, stderr, exit_code, timestamps
- [ ] Write pre/post checks to `job_checks`
- [ ] ACK message after successful DB write; NACK on transient SSH errors

### SSH executor

- [ ] SSH connect (`golang.org/x/crypto/ssh`) — password auth
- [ ] SCP upload config + script to temp path on node
- [ ] Execute remote script; capture stdout/stderr/exit code
- [ ] Pre-run check: SSH reachable (`echo ok`)
- [ ] Connection timeout + reasonable command timeout

### Task dispatch

- [ ] Route by `taskType` to correct bash script
- [ ] Post-run checks per task type
- [ ] Update parent event counts + status when all jobs terminal

---

## Phase 5 — Bash task scripts (`scripts/bash/`)

Platform-owned scripts; embedded or copied at runtime. Each script: `set -euo pipefail`.

- [ ] `deploy.sh` — backup → copy → validate → restart → is-active
- [ ] `validate.sh` — SCP config → `otelcol validate` only
- [ ] `restart.sh` — `systemctl restart otelcol` + is-active
- [ ] `status.sh` — report collector running/stopped
- [ ] `logs.sh` — tail `journalctl -u otelcol`
- [ ] `rollback.sh` — restore `.bak` → validate → restart
- [ ] `stop.sh` — `systemctl stop otelcol`
- [ ] `ssh_test.sh` — `echo ok`

---

## Phase 6 — React UI (`web/`)

### App shell

- [ ] Vite + React + TypeScript + React Router
- [ ] API client with JWT in localStorage / Authorization header
- [ ] Protected routes — redirect to login if unauthenticated
- [ ] Role-aware nav (admin sees admin dashboard link)

### Pages

- [ ] **Login** — email/password form; no register link
- [ ] **Instances** — list + add single form
- [ ] **Bulk add** — CSV paste, preview rows, save
- [ ] **Launch event** — task picker, config upload (when required), node multi-select, launcher fields
- [ ] **Events list** — table with status; admin filter bar
- [ ] **Event detail** — job table, checks, stdout/stderr, poll while RUNNING
- [ ] **Event detail actions** — rollback per job, re-run failed, clone
- [ ] **Admin dashboard** — all events/instances with filters

### UX essentials (internal tool — functional, not fancy)

- [ ] Clear FAILED / PARTIAL / COMPLETED badges
- [ ] Expandable log panel per job row
- [ ] Loading + error states on API calls

---

## Phase 7 — Integration & hardening

### End-to-end flows

- [ ] Login → add instance → SSH test event → success
- [ ] Login → deploy config to 2 nodes → PARTIAL failure visible
- [ ] Re-run failed nodes only
- [ ] Per-node rollback after bad deploy
- [ ] Clone event → edit → launch
- [ ] Bulk CSV import → connectivity test
- [ ] Admin login → filter events by launcher

### Security checklist

- [ ] SSH passwords never in API responses after create
- [ ] Users cannot access other users' instances/events
- [ ] JWT expiry enforced
- [ ] No arbitrary shell from user input — scripts only

### Ops

- [ ] `docker compose up --build` documented in README
- [ ] Env vars documented in `.env.example`
- [ ] RabbitMQ management UI noted (port 15672) for debugging

### Testing (minimal v1)

- [ ] API unit tests: auth, YAML validation, ownership checks
- [ ] Worker unit tests: task routing (mock SSH)
- [ ] Manual test checklist completed (above E2E flows)

---

## Deferred (not v1 — do not track here)

| Item | Target |
|------|--------|
| Rollback all (batch) | v1.1 |
| SSO / LDAP | v2 |
| SSH key auth | v2 |
| Self-service registration | Out of scope |
| Scheduled deploys | v2 |
| AWS SSM executor | v3 |

---

## Quick reference — v1 feature checklist

Cross-check against the design spec:

| Feature | Tracker location |
|---------|------------------|
| Single instance register | Phase 3 + Phase 6 |
| Bulk CSV register | Phase 3 + Phase 6 |
| All 8 predefined tasks | Phase 5 + Phase 4 |
| RabbitMQ async jobs | Phase 2 + Phase 4 |
| Clone event | Phase 3 + Phase 6 |
| Re-run failed nodes | Phase 3 + Phase 6 |
| Per-node rollback | Phase 3 + Phase 5 + Phase 6 |
| User vs admin roles | Phase 2 + Phase 3 + Phase 6 |
| Admin dashboard + filters | Phase 3 + Phase 6 |

---

## Notes

_Add dated notes as you implement:_

- **2026-07-26** — Tracker created. Docs phase complete. Implementation plan: 16 tasks.
