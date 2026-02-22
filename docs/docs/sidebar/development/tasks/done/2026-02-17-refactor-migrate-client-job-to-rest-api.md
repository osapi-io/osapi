---
title: Migrate client job commands from direct NATS to REST API
status: done
created: 2026-02-17
updated: 2026-02-17
completed: 2026-02-17
---

## Objective

All `client job` CLI commands currently bypass the REST API and connect directly
to NATS. This is inconsistent with `client system` and `client network`, which
go through the REST API. Everything under `client` should use the same path: CLI
→ REST API → NATS.

Current state:

```
client system *   → REST API → NATS    (correct)
client network *  → REST API → NATS    (correct)
client job *      → NATS directly      (inconsistent)
```

Target state:

```
client *          → REST API → NATS    (all consistent)
```

## Benefits

- One entry point (API URL) — no NATS host/port needed in CLI
- Auth/middleware applies uniformly
- CLI doesn't need NATS credentials
- Can put load balancer/proxy in front of the API

## Commands to Migrate

All REST API endpoints already exist:

| CLI Command               | Current Path | REST Endpoint                      |
| ------------------------- | ------------ | ---------------------------------- |
| `client job add`          | Direct NATS  | `POST /job`                        |
| `client job list`         | Direct NATS  | `GET /job`                         |
| `client job get`          | Direct NATS  | `GET /job/{id}`                    |
| `client job delete`       | Direct NATS  | `DELETE /job/{id}`                 |
| `client job status`       | Direct NATS  | `GET /job/status`                  |
| `client job run`          | Direct NATS  | `POST /job` + poll `GET /job/{id}` |
| `client job workers list` | Direct NATS  | `GET /job/workers`                 |

## Tasks

- [x] Rewrite each command to use the REST client (`handler`) instead of
      `jobClient`
- [x] Remove NATS setup from `client_job.go` `PersistentPreRun`
- [x] Remove `natsClient`, `jobsKV`, `jobClient` package vars
- [x] Remove NATS-related flags from `client_job.go` (nats-host, nats-port,
      kv-bucket, stream-name, etc.)
- [x] Add any missing REST client handler methods if needed
- [x] Update tests
- [x] Verify `client job run` polling works through REST API

## Notes

- The generated REST client (`internal/client/gen/client.gen.go`) already has
  methods for all endpoints.
- The `client` parent command already creates the REST client in its
  `PersistentPreRun`.
- `client job run` is the most complex — it creates a job then polls for
  completion. This maps to `POST /job` followed by polling `GET /job/{id}`.
