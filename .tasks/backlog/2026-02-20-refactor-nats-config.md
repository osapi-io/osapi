---
title: Refactor NATS config so each consumer owns its connection settings
status: backlog
created: 2026-02-20
updated: 2026-02-20
---

## Objective

The current config puts the API server's NATS connection under `job.client`,
which is confusing — it looks like a job-specific setting when it's really
the API server's NATS connection. Each component that connects to NATS should
declare its own `nats.host` / `nats.port` in its own config section.

## Current Structure (problematic)

```yaml
job:
  client:          # <-- API server's NATS connection, confusingly under "job"
    host: localhost
    port: 4222
    client_name: osapi-jobs-cli
  worker:          # <-- Worker's NATS connection + runtime config mixed together
    host: localhost
    port: 4222
    client_name: osapi-job-worker
    queue_group: job-workers
    hostname: ''
    max_jobs: 10
    labels: {}
```

Also `api.client` is vague — it's the CLI's HTTP client config but "client"
doesn't convey that.

## Proposed Structure

Each component that connects to NATS gets `nats.host` and `nats.port` in
its own section:

```yaml
api:
  server:
    port: 8080
    nats:
      host: localhost
      port: 4222
      client_name: osapi-api
    security:
      signing_key: '<secret>'
      cors:
        allow_origins: []

  cli:                          # renamed from "client" for clarity
    url: 'http://0.0.0.0:8080'
    security:
      bearer_token: '<jwt>'

job:
  worker:
    nats:
      host: localhost
      port: 4222
      client_name: osapi-job-worker
    queue_group: job-workers
    hostname: ''
    max_jobs: 10
    labels: {}
```

## Changes Required

- `internal/config/types.go` — restructure `API`, `JobWorker`, remove
  `JobClient`; rename `Client` to `CLI`
- `cmd/api_server_start.go` — read NATS config from `api.server.nats`
- `cmd/job_worker_start.go` — read NATS config from `job.worker.nats`
- `cmd/client_*.go` — update references from `api.client` to `api.cli`
- `osapi.yaml` — update example config
- `docs/docs/sidebar/configuration.md` — update docs
- Environment variable mapping changes (e.g., `OSAPI_API_CLI_URL`)
- Migration note for existing users

## Notes

- This is a breaking config change — document in release notes
- The `nats.server` section (embedded NATS server) stays as-is
- `job.client` section is removed entirely; its settings move to
  `api.server.nats`
- Consider whether `job.worker.nats` should just be `worker.nats` at
  top level (worker is its own process, not really a sub-concept of job)

## Outcome

_To be filled in when done._
