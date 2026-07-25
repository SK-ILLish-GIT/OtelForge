# OtelForge v1 — Design Spec

**Date:** 2026-07-26  
**Status:** Approved (supersedes stack choices in 2026-07-25 spec)  
**Product:** Internal OTel config deployment tool (~20 operators, ~50 target nodes)

## Overview

**OtelForge** (otel-deploy) is an internal web platform for deploying and managing OpenTelemetry Collector configurations across SSH-reachable Linux nodes. Operators register nodes, launch **events** against one or many targets, and track per-node verification results and logs. Admins audit all activity.

Transport is **SSH + SCP** with **platform-owned scripts** only. No AWS IAM, SSM, or agent on target nodes.

## Architecture

```
React UI → Go API → Postgres
                 ↘ RabbitMQ → Go Worker → SSH/SCP → Linux nodes
```

**Stack:** React (Vite+TS), Go monorepo (API + Worker + shared `internal/`), **PostgreSQL**, RabbitMQ, Docker Compose — **API: Go Fiber v2** (internal VM)

**Repo layout:**

```
OtelForge/
├── docker-compose.yml
├── api/cmd/server/
├── worker/cmd/worker/
├── internal/          # shared: models, crypto, db, queue
├── web/               # React + Vite + TS
└── scripts/bash/      # deploy, rollback, restart, etc.
```

## Audience & Scale

| Dimension | v1 target |
|-----------|-----------|
| Users | ~20 internal operators |
| Target nodes | ~50 per deploy batch |
| Deployment | Single Docker Compose on internal VM |
| Multi-tenancy | None — user/admin roles only |

## Auth & Access

- Email/password login → JWT (24h expiry)
- Passwords: bcrypt in Postgres
- **No self-service registration** — IT/admin seeds accounts (CLI or seed script)
- SSH passwords: AES-256-GCM encrypted at rest; never returned by API after save; worker decrypts in memory only for SSH session

| Role | Instances | Events | Admin views |
|------|-----------|--------|-------------|
| **user** | CRUD own | Create/view own | No |
| **admin** | View all | View all + filters | Yes |

**Audit fields on every event:** `launcherName`, `launcherEmail` (auto-filled from login, editable before submit)

## Core User Flow

1. Login
2. Register instance (host, port, SSH user, password) — or bulk CSV import
3. Launch event → pick task → upload `config.yaml` if required → select instance(s)
4. API validates, creates event + jobs, publishes to RabbitMQ
5. Worker deploys via SSH, runs verification, writes job results to Postgres
6. Track pre/post checks + logs on Event Detail
7. Rollback per node if needed; clone event to re-run; re-run failed nodes only

## v1 Features

| Feature | v1 |
|---------|-----|
| Single instance register | ✓ |
| Bulk CSV register | ✓ |
| All 8 predefined tasks | ✓ |
| RabbitMQ async jobs | ✓ |
| Clone event | ✓ |
| Re-run failed nodes | ✓ |
| Per-node rollback | ✓ |
| User vs admin roles | ✓ |
| Admin dashboard + filters | ✓ |
| Rollback all (batch) | v1.1 |
| SSO / LDAP | v2 |
| SSH key auth | v2 |
| Self-service registration | Out of scope |

## Predefined Tasks (v1)

| Task | Config required? |
|------|------------------|
| Deploy OTel Config | Yes |
| Validate Config Only | Yes |
| Restart Collector | No |
| Check Status | No |
| Fetch Logs | No |
| Rollback Config | No |
| Stop Collector | No |
| SSH Connectivity Test | No |

Scripts are platform-owned (not user shell). Config tasks require `config.yaml` upload.

## Data Model (Postgres)

### Tables

- **users** — id, email, password_hash, role (`user` | `admin`), created_at
- **instances** — id, owner_id, name, host, port, ssh_user, ssh_password_enc, created_at
- **events** — id, owner_id, name, launcher_name, launcher_email, task_type, config_content, config_checksum, status, related_event_id, rollback_scope, total_jobs, verified_count, failed_count, created_at
- **jobs** — id, event_id, instance_id, status, stdout, stderr, exit_code, started_at, finished_at
- **job_checks** — id, job_id, phase (`pre` | `post`), name, passed, message

### Status enums

**Event:** `QUEUED` → `RUNNING` → `COMPLETED` | `PARTIAL` | `FAILED`  
**Job:** `QUEUED` → `RUNNING` → `VERIFIED` | `FAILED`

## Verification

| Phase | Where | Checks |
|-------|-------|--------|
| **Pre-run (API)** | Before queue | Valid YAML; instance ownership |
| **Pre-run (worker)** | Before task | SSH reachable |
| **Post-run (worker)** | After task | Task-specific (validate, restart, active, logs) |

## Rollback

- **Deploy** backs up current config → `config.yaml.bak` before overwrite
- **Per-node:** Rollback button on job row in Event Detail (if backup exists)
- **Rollback all:** deferred to v1.1

## Message Queue

| Element | Value |
|---------|-------|
| Exchange | `otel.deploy` (direct) |
| Main queue | `deploy.jobs` (durable) |
| Dead-letter queue | `deploy.jobs.dlq` |
| Prefetch | 1 per worker |
| Max retries | 3 → DLQ |

Message payload: `{ jobId, eventId, instanceId, taskType }`

## UI Pages (v1)

- Login (no public register)
- Instances (add single + bulk CSV)
- Launch Event (task picker, config upload, node multi-select)
- Events List (user: own; admin: all + filters)
- Event Detail (progress, checks, logs, rollback, re-run failed, clone)
- Admin Dashboard (filter by launcher, status, task, date)

## Out of Scope (v1)

- MongoDB (replaced by Postgres)
- Rollback all batch action
- SSH key auth, AWS SSM executor, custom user scripts, scheduled deploys, onboarding wizard, SSO

## Related Docs

- [Architecture & Product Guide](../../otel-deploy/ARCHITECTURE.md) — full narrative (aligned to this spec)
- [v1 Task Tracker](../plans/2026-07-26-otelforge-v1-task-tracker.md) — checkbox checklist
- [v1 Implementation Plan](../plans/2026-07-26-otelforge-v1-implementation-plan.md) — 16-task build plan
- [Original design spec](./2026-07-25-otel-deploy-design.md) — historical; MongoDB stack superseded
