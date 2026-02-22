---
title: Add NATS authentication and namespace support to config
status: done
created: 2026-02-20
updated: 2026-02-21
---

## Objective

The nats-client library already supports three auth types (NoAuth, UserPass,
NKey), but OSAPI hardcoded `NoAuth` in all three components. Surface these auth
options through `osapi.yaml` and plumb them into the client connections. Also
add namespace/subject prefix support so multiple OSAPI deployments can share a
NATS cluster without colliding.

## Outcome

Implemented full NATS auth and namespace support across all three components:

### Config restructure

- Moved JetStream infrastructure config (`stream`, `kv`, `dlq`) from `job.*` to
  `nats.*` — NATS owns infrastructure, job owns worker runtime
- Moved consumer config from `job.consumer` to `job.worker.consumer`
- Added `namespace` and `auth` fields to `nats.server`, `api.server.nats`, and
  `job.worker.nats`

### Auth support

- Added `NATSAuth` (client-side) and `NATSServerAuth` (server-side) config types
  supporting `none`, `user_pass`, and `nkey` auth
- Created `cmd/nats_auth.go` helper to bridge config types to nats-client
  `AuthOptions`
- Server-side auth configures `natsserver.Options.Users` / `.Nkeys`
- Client-side auth passes credentials through `natsclient.AuthOptions`

### Namespace support

- `job.Init(namespace)` sets global `JobsQueryPrefix` / `JobsModifyPrefix` with
  namespace prefix (e.g., `osapi.jobs.query`)
- `ApplyNamespaceToInfraName()` prefixes stream/KV names (e.g., `osapi-JOBS`,
  `osapi-job-queue`)
- `ApplyNamespaceToSubjects()` prefixes stream subjects (e.g., `osapi.jobs.>`)
- `ParseSubject()` handles namespace-prefixed subjects transparently
- Worker uses namespaced `streamName` passed at construction time

### Files changed

- `internal/config/types.go` — new auth/namespace/stream/kv/dlq types
- `internal/job/subjects.go` — namespace Init, Apply helpers
- `internal/job/config.go` — updated signatures to use new config types
- `internal/job/worker/types.go` — added `streamName` field
- `internal/job/worker/worker.go` — `streamName` parameter in `New()`
- `internal/job/worker/consumer.go` — uses `w.streamName`, dynamic prefixes
- `internal/job/worker/handler.go` — uses dynamic
  `JobsQueryPrefix`/`JobsModifyPrefix`
- `internal/job/client/client.go` — `StreamName` in Options
- `internal/job/client/jobs.go` — dynamic DLQ name from `streamName`
- `cmd/nats_auth.go` — new auth helper
- `cmd/nats_server_start.go` — server auth, namespace, inline infra setup
- `cmd/api_server_start.go` — client auth, namespace
- `cmd/job_worker_start.go` — client auth, namespace, pass streamName
- `cmd/nats_server.go` — debug logging for new fields
- `cmd/api_server.go` — debug logging for NATS connection
- `cmd/job_worker.go` — debug logging, fixed viper bindings
- `configs/osapi.yaml` — restructured config
- `configs/osapi.local.yaml` — restructured config
- `configs/osapi.nerd.yaml` — restructured config
- `test/osapi.yaml` — restructured config
- `docs/docs/sidebar/configuration.md` — full rewrite with auth/namespace docs
- All test files updated for new config types and function signatures
