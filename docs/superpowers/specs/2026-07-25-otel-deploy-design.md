# otel-deploy — Design Spec (Historical)

**Date:** 2026-07-25  
**Status:** Superseded for stack choices — see [2026-07-26 OtelForge v1 Design](./2026-07-26-otelforge-v1-design.md)  
**Full docs:** [Architecture & Product Guide](../../otel-deploy/ARCHITECTURE.md)

> **Note:** This document originally specified MongoDB. The approved v1 implementation uses **PostgreSQL** instead. Product scope, flows, and task definitions remain valid; refer to the 2026-07-26 spec for authoritative v1 decisions.

## Overview

Cloud-agnostic web platform to deploy OpenTelemetry `config.yaml` to any SSH-reachable Linux node (EC2, GCP, Azure, on-prem). No AWS IAM or SSM required. Users register nodes, launch events, track verification/logs. Admins see all activity.

## Architecture (original — MongoDB superseded)

```
React UI → Go API → MongoDB          ← use Postgres per v1 spec
                 ↘ RabbitMQ → Go Worker → SSH/SCP → EC2
```

**Stack:** React (Vite+TS), Go API, Go Worker, ~~MongoDB~~ **PostgreSQL**, RabbitMQ, Docker Compose

## Auth & Access

- Email/password login (JWT)
- **User:** own instances + events only
- **Admin:** all instances, events, logs; filter by launcher email/status/date

## Core User Flow

1. Login
2. Register instance (host, port, SSH user, password — encrypted at rest)
3. Launch event → upload `config.yaml` → select instance(s)
4. Worker deploys config, validates, restarts `otelcol`
5. Track pre/post checks + logs on Event Detail
6. Rollback if needed (per instance or all)

## Additional v1 Flows

| Flow | Summary |
|------|---------|
| **Failed deploy recovery** | View failures → re-run failed instances / rollback one or all / re-deploy fixed config |
| **Clone event** | Copy past event (config, task, instances) → launch new event |
| **Bulk register** | CSV paste `name,host,port,user,password` → save all → optional connectivity test |
| **Admin audit** | Browse/filter all events by launcher, status, task, date |

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

## Event Model

**Event holds:** `name`, `launcherName`, `launcherEmail`, `taskType`, `configContent?`, `instanceIds[]`, `status`, job counts, `relatedEventId?`, `rollbackScope?`

**Job holds (per instance):** `status`, `preRunChecks[]`, `postRunChecks[]`, `stdout`, `stderr`, `exitCode`

**Event status:** `QUEUED` → `RUNNING` → `COMPLETED` | `PARTIAL` | `FAILED`  
**Job status:** `QUEUED` → `RUNNING` → `VERIFIED` | `FAILED`

## Verification

| Phase | Checks |
|-------|--------|
| **Pre-run (API)** | Valid YAML; instance ownership |
| **Pre-run (worker)** | SSH reachable |
| **Post-run (worker)** | `otelcol validate`, restart, `systemctl is-active`, logs |

## Rollback

- **Deploy** backs up current config → `config.yaml.bak` before applying new one
- **Per instance:** Rollback button on job row (if backup exists)
- **Whole event:** Rollback all eligible instances from original deploy; skip instances with no backup

## Pages (v1)

- Login / Register
- Instances (add + bulk add)
- Launch Event
- Events list (admin: all + filters)
- Event Detail (logs, rollback, re-run failed, clone)

## Positioning

- **Not** an AWS SSM replacement — OTel-specific deploy/audit/rollback tool
- **Not** AWS-only — any node with SSH + otelcol
- **No IAM** — worker connects via SSH to host/IP, not AWS API

## Out of Scope (v1)

- SSH key auth, AWS SSM executor, custom user scripts, scheduled deploys, onboarding wizard
