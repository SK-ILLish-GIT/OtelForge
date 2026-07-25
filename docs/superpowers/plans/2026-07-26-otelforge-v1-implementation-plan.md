# OtelForge v1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an internal web platform to deploy OpenTelemetry `config.yaml` to SSH-reachable Linux nodes with async jobs, verification, rollback, and admin audit.

**Architecture:** React SPA talks to a Go REST API that persists to PostgreSQL and enqueues per-node jobs on RabbitMQ. Go workers consume messages, SSH/SCP to targets, run platform-owned bash scripts, and write job results back to Postgres.

**Tech Stack:** Go 1.22+, PostgreSQL 16, RabbitMQ 3, React 18, Vite, TypeScript, **Go Fiber v2**, jackc/pgx, amqp091-go, golang.org/x/crypto/ssh, golang-jwt/jwt, Docker Compose

**Spec:** [2026-07-26-otelforge-v1-design.md](../specs/2026-07-26-otelforge-v1-design.md)  
**Tracker:** [2026-07-26-otelforge-v1-task-tracker.md](./2026-07-26-otelforge-v1-task-tracker.md)

## Global Constraints

- Internal tool: ~20 operators, ~50 target nodes per batch; Docker Compose on internal VM
- PostgreSQL only (no MongoDB)
- RabbitMQ exchange `otel.deploy`, queue `deploy.jobs`, DLQ `deploy.jobs.dlq`, prefetch 1, max 3 retries
- JWT auth, 24h expiry; bcrypt user passwords; AES-256-GCM for SSH passwords at rest
- Roles: `user` (own resources) and `admin` (all resources + filters)
- No self-service registration; seed users via CLI
- No arbitrary user shell — only embedded scripts in `scripts/bash/`
- Event status: `QUEUED` → `RUNNING` → `COMPLETED` | `PARTIAL` | `FAILED`
- Job status: `QUEUED` → `RUNNING` → `VERIFIED` | `FAILED`
- Config path on nodes: `/etc/otelcol/config.yaml`, backup: `config.yaml.bak`
- API never returns decrypted SSH passwords after create/update

## File Map

| Path | Responsibility |
|------|----------------|
| `go.mod` | Monorepo module `github.com/otelforge/otelforge` |
| `internal/config/config.go` | Load env: DATABASE_URL, RABBITMQ_URL, JWT_SECRET, ENCRYPTION_KEY |
| `internal/models/models.go` | User, Instance, Event, Job, JobCheck, TaskType, status enums |
| `internal/db/migrate.go` | Run SQL migrations on startup |
| `internal/db/migrations/001_initial.sql` | Schema |
| `internal/db/store.go` | CRUD queries (pgx) |
| `internal/crypto/password.go` | bcrypt hash/verify |
| `internal/crypto/secretbox.go` | AES-256-GCM encrypt/decrypt |
| `internal/queue/rabbitmq.go` | Declare topology, Publish, Consume |
| `internal/queue/message.go` | JobMessage JSON type |
| `internal/auth/jwt.go` | Issue/parse JWT, context user |
| `internal/auth/middleware.go` | RequireAuth, RequireAdmin |
| `internal/ssh/client.go` | Connect, SCP, Run script |
| `internal/tasks/scripts.go` | Embed `scripts/bash/*.sh`, resolve by TaskType |
| `internal/events/service.go` | Create event, enqueue jobs, rollup status |
| `internal/yamlutil/validate.go` | YAML syntax check |
| `api/cmd/server/main.go` | HTTP server, route registration |
| `api/internal/handlers/*.go` | auth, instances, events, admin handlers |
| `worker/cmd/worker/main.go` | Queue consumer loop |
| `worker/internal/runner/runner.go` | Job execution pipeline |
| `cmd/seed/main.go` | Seed admin user |
| `scripts/bash/*.sh` | 8 remote task scripts |
| `web/src/*` | React UI |
| `docker-compose.yml` | Postgres, RabbitMQ, api, worker, web |

---

### Task 1: Monorepo scaffold and Docker

**Files:**
- Create: `go.mod`, `api/cmd/server/main.go`, `worker/cmd/worker/main.go`, `api/Dockerfile`, `worker/Dockerfile`, `web/package.json`, `web/vite.config.ts`, `web/Dockerfile`, `README.md`
- Modify: `docker-compose.yml` (already exists — verify env vars)

**Interfaces:**
- Produces: runnable `go run ./api/cmd/server` and `go run ./worker/cmd/worker` stubs returning 200/health or log "started"

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/sahil.sk/Desktop/Me/ColdMail-Github/Untitled/OtelForge
go mod init github.com/otelforge/otelforge
```

- [ ] **Step 2: Create API stub**

Create `api/cmd/server/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	addr := envOr("API_ADDR", ":8080")
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	log.Printf("api listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
```

- [ ] **Step 3: Create worker stub**

Create `worker/cmd/worker/main.go`:

```go
package main

import "log"

func main() {
	log.Println("worker started (stub)")
	select {}
}
```

- [ ] **Step 4: Verify compile**

Run: `go build -o /dev/null ./api/cmd/server && go build -o /dev/null ./worker/cmd/worker`  
Expected: exit 0

- [ ] **Step 5: Scaffold web**

Run: `npm create vite@latest web -- --template react-ts`  
Update `web/vite.config.ts` proxy if needed for local API.

- [ ] **Step 6: Add Dockerfiles and README**

Multi-stage Go Dockerfiles copying `go.mod`, `internal/`, `api/` or `worker/`. README with `docker compose up --build` and `.env.example` copy instructions.

- [ ] **Step 7: Commit**

```bash
git add go.mod api/ worker/ web/ README.md api/Dockerfile worker/Dockerfile web/Dockerfile
git commit -m "chore: scaffold monorepo, api/worker stubs, web, dockerfiles"
```

---

### Task 2: Config and domain models

**Files:**
- Create: `internal/config/config.go`, `internal/models/models.go`, `internal/config/config_test.go`
- Test: `internal/config/config_test.go`, `internal/models/models_test.go`

**Interfaces:**
- Produces: `config.Load() (Config, error)`, model types `User`, `Instance`, `Event`, `Job`, `JobCheck`, `TaskType`, `Role`, `EventStatus`, `JobStatus`

- [ ] **Step 1: Write failing config test**

Create `internal/config/config_test.go`:

```go
func TestLoad_requiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL missing")
	}
}
```

- [ ] **Step 2: Run test — expect FAIL**

Run: `go test ./internal/config/... -v -run TestLoad_requiresDatabaseURL`

- [ ] **Step 3: Implement config**

Create `internal/config/config.go`:

```go
package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	RabbitMQURL       string
	JWTSecret         string
	EncryptionKey     string // 32 bytes hex or raw
	APIAddr           string
	CORSOrigin        string
	WorkerConcurrency int
}

func Load() (Config, error) {
	db := os.Getenv("DATABASE_URL")
	if db == "" {
		return Config{}, fmt.Errorf("DATABASE_URL required")
	}
	wc, _ := strconv.Atoi(envOr("WORKER_CONCURRENCY", "2"))
	return Config{
		DatabaseURL:       db,
		RabbitMQURL:       envOr("RABBITMQ_URL", "amqp://otel:otel@localhost:5672/"),
		JWTSecret:         envOr("JWT_SECRET", "dev-jwt-secret-change-in-production"),
		EncryptionKey:     os.Getenv("ENCRYPTION_KEY"),
		APIAddr:           envOr("API_ADDR", ":8080"),
		CORSOrigin:        envOr("CORS_ORIGIN", "http://localhost:5173"),
		WorkerConcurrency: wc,
	}, nil
}
```

- [ ] **Step 4: Implement models**

Create `internal/models/models.go` with:

```go
type Role string
const RoleUser Role = "user"; const RoleAdmin Role = "admin"

type TaskType string
const (
	TaskDeployConfig   TaskType = "deploy_config"
	TaskValidateConfig TaskType = "validate_config"
	TaskRestart        TaskType = "restart_collector"
	TaskCheckStatus    TaskType = "check_status"
	TaskFetchLogs      TaskType = "fetch_logs"
	TaskRollback       TaskType = "rollback_config"
	TaskStop           TaskType = "stop_collector"
	TaskSSHTest        TaskType = "ssh_connectivity_test"
)

func (t TaskType) RequiresConfig() bool {
	return t == TaskDeployConfig || t == TaskValidateConfig
}

type EventStatus string
const (
	EventQueued    EventStatus = "QUEUED"
	EventRunning   EventStatus = "RUNNING"
	EventCompleted EventStatus = "COMPLETED"
	EventPartial   EventStatus = "PARTIAL"
	EventFailed    EventStatus = "FAILED"
)

type JobStatus string
const (
	JobQueued   JobStatus = "QUEUED"
	JobRunning  JobStatus = "RUNNING"
	JobVerified JobStatus = "VERIFIED"
	JobFailed   JobStatus = "FAILED"
)
```

Plus struct fields matching spec (UUID ids as `string` or `uuid.UUID` — pick `github.com/google/uuid`).

- [ ] **Step 5: Run tests — expect PASS**

Run: `go test ./internal/config/... ./internal/models/... -v`

- [ ] **Step 6: Commit**

```bash
git commit -am "feat: add config loader and domain models"
```

---

### Task 3: PostgreSQL schema and store

**Files:**
- Create: `internal/db/migrations/001_initial.sql`, `internal/db/migrate.go`, `internal/db/store.go`, `internal/db/store_test.go`

**Interfaces:**
- Consumes: `config.Config`, `models.*`
- Produces: `db.Connect(ctx, cfg)`, `db.Migrate(ctx)`, `Store` methods: `CreateUser`, `GetUserByEmail`, `CreateInstance`, `ListInstances`, `CreateEvent`, `CreateJob`, `UpdateJob`, `GetEventWithJobs`, `ListEvents(filter)`, etc.

- [ ] **Step 1: Write migration SQL**

Create `internal/db/migrations/001_initial.sql`:

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('user', 'admin')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE instances (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id UUID NOT NULL REFERENCES users(id),
  name TEXT NOT NULL,
  host TEXT NOT NULL,
  port INT NOT NULL DEFAULT 22,
  ssh_user TEXT NOT NULL,
  ssh_password_enc BYTEA NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id UUID NOT NULL REFERENCES users(id),
  name TEXT NOT NULL,
  launcher_name TEXT NOT NULL,
  launcher_email TEXT NOT NULL,
  task_type TEXT NOT NULL,
  config_content TEXT,
  config_checksum TEXT,
  status TEXT NOT NULL,
  related_event_id UUID REFERENCES events(id),
  rollback_scope TEXT,
  total_jobs INT NOT NULL DEFAULT 0,
  verified_count INT NOT NULL DEFAULT 0,
  failed_count INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_launcher_email ON events(launcher_email);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_created_at ON events(created_at);
CREATE INDEX idx_events_task_type ON events(task_type);

CREATE TABLE jobs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  instance_id UUID NOT NULL REFERENCES instances(id),
  status TEXT NOT NULL,
  stdout TEXT,
  stderr TEXT,
  exit_code INT,
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ
);

CREATE TABLE job_checks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  phase TEXT NOT NULL CHECK (phase IN ('pre', 'post')),
  name TEXT NOT NULL,
  passed BOOLEAN NOT NULL,
  message TEXT NOT NULL
);

CREATE TABLE event_instances (
  event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  instance_id UUID NOT NULL REFERENCES instances(id),
  PRIMARY KEY (event_id, instance_id)
);
```

- [ ] **Step 2: Implement migrate + store with pgx**

Add dependency: `go get github.com/jackc/pgx/v5/pgxpool`

Implement `Migrate` reading embedded SQL; `Store` with parameterized queries.

- [ ] **Step 3: Integration test with testcontainers or skip if no docker**

Minimal test: `TestMigrate_idempotent` against `DATABASE_URL` when set.

- [ ] **Step 4: Commit**

```bash
git commit -am "feat: postgres schema and store layer"
```

---

### Task 4: Crypto (bcrypt + AES-GCM)

**Files:**
- Create: `internal/crypto/password.go`, `internal/crypto/secretbox.go`, `internal/crypto/secretbox_test.go`

**Interfaces:**
- Produces: `HashPassword(plain string) (string, error)`, `CheckPassword(hash, plain string) error`, `Encrypt(plaintext, key []byte) ([]byte, error)`, `Decrypt(ciphertext, key []byte) ([]byte, error)`

- [ ] **Step 1: Write failing roundtrip test**

```go
func TestEncryptDecrypt_roundtrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	ct, err := Encrypt([]byte("secret"), key)
	if err != nil { t.Fatal(err) }
	pt, err := Decrypt(ct, key)
	if err != nil { t.Fatal(err) }
	if string(pt) != "secret" { t.Fatalf("got %q", pt) }
}
```

- [ ] **Step 2: Implement AES-256-GCM**

Use `crypto/aes` + `cipher.NewGCM`; prepend nonce to ciphertext.

- [ ] **Step 3: Implement bcrypt wrappers**

Use `golang.org/x/crypto/bcrypt` cost 12.

- [ ] **Step 4: Run tests — PASS**

Run: `go test ./internal/crypto/... -v`

- [ ] **Step 5: Commit**

```bash
git commit -am "feat: bcrypt passwords and AES-GCM secret encryption"
```

---

### Task 5: RabbitMQ queue layer

**Files:**
- Create: `internal/queue/message.go`, `internal/queue/rabbitmq.go`, `internal/queue/rabbitmq_test.go`

**Interfaces:**
- Produces: `type JobMessage struct { JobID, EventID, InstanceID, TaskType string }`, `Queue.Declare(ctx)`, `Queue.Publish(ctx, JobMessage)`, `Queue.Consume(ctx, handler func(JobMessage) error)`

- [ ] **Step 1: Add dependency**

```bash
go get github.com/rabbitmq/amqp091-go
```

- [ ] **Step 2: Implement topology**

Exchange: `otel.deploy` direct, durable. Queue: `deploy.jobs` bound with routing key `job`. DLQ: `deploy.jobs.dlq`. Set `x-dead-letter-exchange` and max retry via header or republish count.

- [ ] **Step 3: Publish/consume JSON messages**

```go
type JobMessage struct {
	JobID      string `json:"jobId"`
	EventID    string `json:"eventId"`
	InstanceID string `json:"instanceId"`
	TaskType   string `json:"taskType"`
}
```

Consumer: QoS prefetch=1; ACK after handler returns nil; NACK requeue on transient error.

- [ ] **Step 4: Manual smoke test**

With RabbitMQ running: publish one message, consume in test main.

- [ ] **Step 5: Commit**

```bash
git commit -am "feat: rabbitmq queue with DLQ topology"
```

---

### Task 6: Bash task scripts (all 8)

**Files:**
- Create: `scripts/bash/deploy.sh`, `validate.sh`, `restart.sh`, `status.sh`, `logs.sh`, `rollback.sh`, `stop.sh`, `ssh_test.sh`
- Create: `internal/tasks/scripts.go` (embed.FS)

**Interfaces:**
- Produces: `tasks.ScriptFor(TaskType) ([]byte, error)`, `tasks.ScriptName(TaskType) string`

- [ ] **Step 1: Write deploy.sh** (from ARCHITECTURE.md)

```bash
#!/usr/bin/env bash
set -euo pipefail
CONFIG_SRC="$1"
CONFIG_DEST="/etc/otelcol/config.yaml"
BACKUP="${CONFIG_DEST}.bak"
if [ -f "$CONFIG_DEST" ]; then sudo cp "$CONFIG_DEST" "$BACKUP"; fi
sudo cp "$CONFIG_SRC" "$CONFIG_DEST"
sudo otelcol validate --config="$CONFIG_DEST"
sudo systemctl restart otelcol
sudo systemctl is-active otelcol
```

- [ ] **Step 2: Write remaining 7 scripts**

`validate.sh`, `restart.sh`, `status.sh`, `logs.sh` (journalctl -u otelcol --no-pager -n 200), `rollback.sh`, `stop.sh`, `ssh_test.sh` (echo ok).

- [ ] **Step 3: Embed in Go**

```go
//go:embed ../../scripts/bash/*.sh
var scriptFS embed.FS
```

- [ ] **Step 4: Commit**

```bash
git commit -am "feat: add eight platform-owned bash task scripts"
```

---

### Task 7: Auth (JWT + seed CLI)

**Files:**
- Create: `internal/auth/jwt.go`, `internal/auth/middleware.go`, `internal/auth/context.go`, `cmd/seed/main.go`, `api/internal/handlers/auth.go`

**Interfaces:**
- Produces: `auth.IssueToken(user models.User) (string, error)`, `auth.UserFromContext(ctx) (models.User, bool)`, middleware `RequireAuth`, `RequireAdmin`

- [ ] **Step 1: Write failing JWT test**

Parse token returns correct `userID` and `role`.

- [ ] **Step 2: Implement JWT (24h expiry)**

Use `github.com/golang-jwt/jwt/v5`, HS256, claims: `sub`, `email`, `role`, `exp`.

- [ ] **Step 3: Implement seed command**

`cmd/seed/main.go`: flags `--email`, `--password`, `--role admin`; hash password; insert user.

Run: `go run ./cmd/seed --email admin@internal.local --password changeme --role admin`

- [ ] **Step 4: Implement POST /api/v1/auth/login**

Request: `{ "email", "password" }`. Response: `{ "token", "user": { id, email, role } }`.

- [ ] **Step 5: Commit**

```bash
git commit -am "feat: JWT auth, middleware, login handler, seed CLI"
```

---

### Task 8: Instances API (single + bulk CSV)

**Files:**
- Create: `api/internal/handlers/instances.go`, `api/internal/handlers/instances_test.go`

**Interfaces:**
- Consumes: `Store`, `crypto.Encrypt`, `auth middleware`
- Produces: REST routes under `/api/v1/instances`

- [ ] **Step 1: Write failing test — user cannot read other's instance**

- [ ] **Step 2: Implement CRUD handlers**

Encrypt SSH password on create/update; never include `ssh_password_enc` in JSON responses.

- [ ] **Step 3: Implement bulk CSV**

`POST /api/v1/instances/bulk` body: `{ "csv": "name,host,port,username,password\n..." }`  
Parse header row; validate; insert all; return `{ "created": N, "errors": [] }`.

- [ ] **Step 4: Admin list all instances**

Admin role returns all with `ownerEmail`; user returns own only.

- [ ] **Step 5: Commit**

```bash
git commit -am "feat: instances API with bulk CSV import"
```

---

### Task 9: Events API (create, list, detail, clone, rerun, rollback)

**Files:**
- Create: `internal/yamlutil/validate.go`, `internal/events/service.go`, `api/internal/handlers/events.go`

**Interfaces:**
- Consumes: `Store`, `Queue.Publish`, `yamlutil.ValidateYAML`
- Produces: `events.Create(ctx, input CreateEventInput) (*models.Event, error)` — creates event, jobs, publishes messages

- [ ] **Step 1: YAML validation test**

```go
func TestValidateYAML_rejectsInvalid(t *testing.T) {
	err := ValidateYAML([]byte(":\n  bad"))
	if err == nil { t.Fatal("expected error") }
}
```

Use `gopkg.in/yaml.v3` unmarshal only (syntax check).

- [ ] **Step 2: Implement CreateEvent**

Validate ownership of all `instanceIds`; if `taskType.RequiresConfig()` require non-empty YAML; set status `QUEUED`; create one job per instance; publish `JobMessage` each; set `total_jobs`.

- [ ] **Step 3: Implement GET list with admin filters**

Query params: `launcherEmail`, `status`, `taskType`, `from`, `to`.

- [ ] **Step 4: Implement GET /events/:id**

Return event + jobs + nested `checks` per job.

- [ ] **Step 5: Implement clone**

`POST /events/:id/clone` → new event with copied task, config, instance list; user edits name before launch optional via body override.

- [ ] **Step 6: Implement rerun-failed**

Reset `FAILED` jobs to `QUEUED`, re-publish messages; event → `RUNNING`.

- [ ] **Step 7: Implement per-job rollback**

`POST /events/:id/jobs/:jobId/rollback` creates new event with `taskType=rollback_config`, `relatedEventId`, single instance.

- [ ] **Step 8: Implement event status rollup helper**

When all jobs terminal: all VERIFIED → COMPLETED; all FAILED → FAILED; mix → PARTIAL.

- [ ] **Step 9: Commit**

```bash
git commit -am "feat: events API with enqueue, clone, rerun, rollback"
```

---

### Task 10: Admin API

**Files:**
- Create: `api/internal/handlers/admin.go`

**Interfaces:**
- Consumes: `RequireAdmin`, `Store.ListEvents`, `Store.ListInstances`

- [ ] **Step 1: GET /api/v1/admin/events** — same filters as list, all owners

- [ ] **Step 2: GET /api/v1/admin/instances** — include owner email

- [ ] **Step 3: Commit**

```bash
git commit -am "feat: admin dashboard API endpoints"
```

---

### Task 11: Worker — SSH executor

**Files:**
- Create: `internal/ssh/client.go`, `internal/ssh/client_test.go`

**Interfaces:**
- Produces: `ssh.RunTask(ctx, ssh.Config, script []byte, configContent *string) (stdout, stderr string, exitCode int, err error)`

- [ ] **Step 1: Implement password auth SSH client**

Use `golang.org/x/crypto/ssh` + `github.com/pkg/sftp` or scp via session `scp -t`.

- [ ] **Step 2: Pre-run check method**

`RunRemote(ctx, "echo ok")` with 15s timeout.

- [ ] **Step 3: Upload script + optional config to `/tmp/otelforge/`**

Execute `bash /tmp/otelforge/script.sh [/tmp/otelforge/config.yaml]`.

- [ ] **Step 4: Commit**

```bash
git commit -am "feat: SSH/SCP executor for remote tasks"
```

---

### Task 12: Worker — job runner and consumer

**Files:**
- Create: `worker/internal/runner/runner.go`, update `worker/cmd/worker/main.go`

**Interfaces:**
- Consumes: `Queue.Consume`, `Store`, `ssh.RunTask`, `tasks.ScriptFor`, `crypto.Decrypt`

- [ ] **Step 1: Implement runner pipeline**

1. Mark job RUNNING  
2. Pre-check SSH  
3. Load script by TaskType  
4. Run script  
5. Post-checks (parse exit code, is-active output)  
6. Mark VERIFIED or FAILED; write checks  
7. Rollup event status  
8. ACK message  

- [ ] **Step 2: Wire worker main**

Load config, connect db, declare queue, start `WorkerConcurrency` goroutines consuming.

- [ ] **Step 3: Manual E2E against test VM or mock**

Deploy to localhost SSH if available; otherwise document manual test in README.

- [ ] **Step 4: Commit**

```bash
git commit -am "feat: worker job runner consuming rabbitmq"
```

---

### Task 13: React app shell + login

**Files:**
- Create: `web/src/api/client.ts`, `web/src/auth/AuthContext.tsx`, `web/src/pages/Login.tsx`, `web/src/App.tsx`, `web/src/components/ProtectedRoute.tsx`

**Interfaces:**
- Produces: `api.login(email, password)`, token stored in `localStorage`, axios/fetch wrapper with `Authorization: Bearer`

- [ ] **Step 1: API client with base URL from `import.meta.env.VITE_API_URL`**

- [ ] **Step 2: Login page — no register link**

- [ ] **Step 3: Protected routes redirect to /login**

- [ ] **Step 4: Commit**

```bash
git commit -am "feat: react auth shell and login page"
```

---

### Task 14: React — instances and bulk CSV

**Files:**
- Create: `web/src/pages/Instances.tsx`, `web/src/pages/BulkAddInstances.tsx`

- [ ] **Step 1: Instances list + add form**

- [ ] **Step 2: Bulk CSV paste + preview table + submit**

- [ ] **Step 3: Commit**

```bash
git commit -am "feat: instances and bulk CSV UI"
```

---

### Task 15: React — events, detail, admin

**Files:**
- Create: `web/src/pages/Events.tsx`, `web/src/pages/EventDetail.tsx`, `web/src/pages/LaunchEvent.tsx`, `web/src/pages/AdminDashboard.tsx`

- [ ] **Step 1: Launch event — task select, config upload when required, multi-select nodes**

- [ ] **Step 2: Events list with status badges; admin filter bar**

- [ ] **Step 3: Event detail — poll every 3s while RUNNING; job table with logs; rollback / rerun / clone buttons**

- [ ] **Step 4: Admin dashboard — tabs or sections for all events/instances**

- [ ] **Step 5: Commit**

```bash
git commit -am "feat: events UI, detail actions, admin dashboard"
```

---

### Task 16: Integration, README, compose health

**Files:**
- Modify: `README.md`, `docker-compose.yml`, `api/cmd/server/main.go` (run migrate on startup)

- [ ] **Step 1: API runs migrations on boot**

- [ ] **Step 2: API declares RabbitMQ topology on boot**

- [ ] **Step 3: Full stack smoke**

Run: `docker compose up --build`  
Expected: all services healthy; login via seeded admin; create instance; launch SSH test event.

- [ ] **Step 4: Document manual E2E checklist** (from task tracker Phase 7)

- [ ] **Step 5: Update task tracker checkboxes**

Mark completed items in `2026-07-26-otelforge-v1-task-tracker.md`.

- [ ] **Step 6: Commit**

```bash
git commit -am "chore: integration docs and compose startup wiring"
```

---

## Spec Coverage Self-Review

| Spec requirement | Task |
|------------------|------|
| Postgres data model | Task 3 |
| RabbitMQ topology | Task 5 |
| 8 predefined tasks | Task 6, 12 |
| JWT + roles | Task 7 |
| No self-service register | Task 7 (seed only) |
| AES SSH encryption | Task 4 |
| Instances single + bulk | Task 8, 14 |
| Events create/list/detail | Task 9, 15 |
| Clone / rerun / rollback | Task 9, 15 |
| Admin filters | Task 10, 15 |
| Verification phases | Task 9 (API YAML), Task 12 (worker checks) |
| UI pages | Tasks 13–15 |
| Docker Compose deploy | Tasks 1, 16 |

**Gaps:** None identified. Rollback-all batch explicitly deferred v1.1 per spec.

**Placeholder scan:** No TBD/TODO in task steps.

**Type consistency:** `TaskType`, `JobMessage`, `JobStatus`, `EventStatus` defined in Task 2 and reused throughout.

---

## Execution Order

```
Task 1 → 2 → 3 → 4 → 5 → 6
                    ↓
         7 → 8 → 9 → 10  (API — sequential)
                    ↓
              11 → 12     (Worker — after queue + scripts + store)
                    ↓
         13 → 14 → 15     (UI — after API endpoints exist)
                    ↓
              16          (Integration)
```

Tasks 6 (scripts) can parallelize with Task 5. UI tasks can start after Task 7 (login) with mocked API if needed.
