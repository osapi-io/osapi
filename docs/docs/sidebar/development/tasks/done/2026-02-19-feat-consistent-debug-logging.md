---
title: Consistent debug logging across all components
status: done
created: 2026-02-19
updated: 2026-02-19
---

## Objective

Make `--debug` output useful and consistent across every component in the
platform. Currently, logging is sparse and inconsistent: the HTTP client layer
has zero request/response logging, NATS KV operations are silent, API handlers
rely entirely on slog-echo middleware with no domain-level logging, and
providers only log warnings for permission issues. This task standardizes
debug-level logging so that running with `--debug` produces a coherent, readable
trace of what the system is doing.

## Scope

### HTTP Client Layer (`internal/client/`)

- Log request method, URL, and response status on every call
- Log request duration at debug level
- Log error responses with body summary

### NATS Client (`nats-client` sibling repo)

- Log individual KV Get/Put/Delete operations with bucket and key
- Log stream Publish with subject
- Log connection events (connect, reconnect, disconnect)

### API Handlers (`internal/api/`)

- Add debug-level log on handler entry with operation name
- Log operation completion with duration
- Use the request logger from echo context

### Job Client (`internal/job/client/`)

- Log all KV reads and writes with job ID and operation
- Log notification publishes with subject

### Worker (`internal/job/worker/`)

- Log provider execution start/finish with operation and duration
- Log result writes to KV

### CLI Commands (`cmd/`)

- Consistent debug logging pattern for all client commands
- Log the endpoint being called and response status

### Standardized Attributes

Define a common set of slog attributes used everywhere:

- `component` — which subsystem (api, worker, client, nats)
- `operation` — what action (kv.get, kv.put, http.request, etc.)
- `job_id` — when operating on a job
- `duration` — for timed operations

## Notes

- All new logging should be at `slog.LevelDebug` so it only appears when
  `--debug` is enabled
- Follow existing slog patterns in the codebase (structured key-value)
- The `nats-client` repo changes need to be coordinated since it's a sibling
  module linked via `replace` in `go.mod`
- Avoid logging sensitive data (tokens, signing keys)
- Keep log lines concise — one line per operation, not multi-line dumps

## Outcome

Implemented across 49 files in 4 phases:

- HTTP client transport logging with method, URL, status, and duration
- NATS KV operation logging (get, put, keys, delete)
- Worker processor dispatch logging (category, operation, provider)
- API handler domain-context logging (routing, job IDs, interfaces)
- Fixed bug: health_status_get.go used package-level slog instead of injected
  logger
- All logging is Debug-level, invisible unless `--debug` is set
