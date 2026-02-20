---
title: Add job retry command
status: backlog
created: 2026-02-19
updated: 2026-02-19
---

## Objective

Add a `job retry --job-id <id>` command that re-submits a stuck or
failed job by reading its stored operation data from KV and publishing
a new notification to the stream. This avoids requiring users to
manually reconstruct and re-run the original CLI command.

## Approach

1. Read the original job's operation data and target from KV
   (`jobs.<id>`)
2. Create a new job with the same operation and target using the
   existing `CreateJob` flow
3. Return the new job ID so the user can track it

The original job remains unchanged (preserves history). The retry
creates a brand new job with a new ID.

## Scope

### Job Client

- `internal/job/client/types.go` — Add `RetryJob` to `JobClient`
  interface
- `internal/job/client/jobs.go` — Implement `RetryJob(ctx, jobID)`
  that reads the original job, extracts operation + target, calls
  `CreateJob`

### API

- `internal/api/job/gen/api.yaml` — Add `POST /job/{id}/retry`
  endpoint
- `internal/api/job/job_retry.go` — Handler that calls
  `RetryJob` and returns the new job ID

### Client Wrapper

- `internal/client/job_retry.go` — `PostJobRetry(ctx, id)` wrapper
- `internal/client/handler.go` — Add to `JobHandler` interface

### CLI

- `cmd/client_job_retry.go` — `job retry --job-id <id>` command
  that calls the API and displays the new job ID

### Tests

- Unit tests for `RetryJob`, API handler, client wrapper

## Notes

- Should work for any terminal status: `submitted` (stuck),
  `failed`, `partial_failure`
- Consider whether to allow retrying `completed` jobs (probably yes,
  for re-running queries)
- The new job gets its own timeline and status events — no link back
  to the original job (keep it simple for now)
- Future enhancement: add `retried_from` field to link job lineage

## Outcome

_To be filled in when done._
