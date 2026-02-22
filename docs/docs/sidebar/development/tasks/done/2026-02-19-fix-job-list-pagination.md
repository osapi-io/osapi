---
title: Add server-side pagination to job list
status: done
created: 2026-02-19
updated: 2026-02-19
---

## Objective

`ListJobs` fetches every key from NATS KV and then calls `GetJobStatus` for each
one, creating an N+1 query problem. With thousands of jobs this is unacceptably
slow. Add server-side pagination so only the requested page of jobs is
processed.

## Current Behavior

1. `ListJobs` calls `kv.Keys()` which returns ALL keys in the bucket
2. For each `jobs.*` key it calls `GetJobStatus()`, which itself calls
   `kv.Keys()` again and reads every status event for that job
3. The CLI's `--limit` and `--offset` flags slice the result client-side after
   everything has already been fetched

With N jobs, this is: 1 `Keys()` + N x (1 `Get` + 1 `Keys` + M `Get`s for status
events). At scale this is O(N\*M) KV reads.

## Proposed Fix

### Phase 1: Server-side limit/offset in ListJobs

Add `limit` and `offset` parameters to `ListJobs()` so it stops calling
`GetJobStatus()` after collecting enough results. This doesn't fix the initial
`Keys()` call but eliminates the N+1 amplification.

- `internal/job/client/jobs.go` — Add `limit`, `offset` params to `ListJobs()`.
  Count matching `jobs.*` keys, skip `offset` entries, stop after `limit`
  results. Return total count alongside results.
- `internal/api/job/job_list.go` — Pass limit/offset from query params to
  `ListJobs()`
- `internal/api/job/gen/api.yaml` — Add `limit` and `offset` query params to the
  `/job` endpoint spec
- `internal/client/client.go` / CLI — Pass limit/offset through the client to
  the API

### Phase 2: Optimize GetJobStatus key scanning

`GetJobStatus` calls `kv.Keys()` to find status events, scanning ALL keys in the
bucket every time. This is the inner loop of the N+1.

Options:

- **Prefix-filtered key listing**: If NATS KV supports `KeysFilter` or similar,
  use `status.{jobID}.*` prefix to avoid scanning unrelated keys
- **Denormalized status field**: Store computed status directly on the job entry
  (updated on each event write) so `ListJobs` can read status from the job entry
  itself without scanning events

### Phase 3: Consider KV bucket separation

Currently jobs, status events, and responses all share one KV bucket. `Keys()`
returns everything. Separating into dedicated buckets (one for job definitions,
one for status events) would make key scanning much cheaper.

## Out of Scope

- Cursor-based pagination (would need NATS KV Watch or similar)
- Full-text search / filtering beyond status
- Database migration (staying with NATS KV)

## Acceptance Criteria

- `job list --limit 10` only processes 10 jobs server-side
- `job list --limit 10 --offset 20` skips 20, processes 10
- Total count is returned for UI summary without processing all jobs
- Existing tests updated, no coverage regression

## Outcome

Phase 1 completed. Phase 2 partially addressed (GetQueueStats N+1 fix).

### Changes made

- **`internal/job/client/types.go`** — Added `ListJobsResult` struct, updated
  `ListJobs` signature to accept `limit`/`offset` and return `*ListJobsResult`
- **`internal/job/client/jobs.go`** — Rewrote `ListJobs` with single `kv.Keys()`
  call, newest-first ordering, server-side pagination. Added
  `getJobStatusFromKeys` helper that reuses pre-fetched keys. Fixed
  `GetQueueStats` N+1 bug (removed redundant inner `kv.Keys()`).
- **`internal/api/job/gen/api.yaml`** — Added `limit` and `offset` query params
  to `GET /job`
- **`internal/api/job/job_list.go`** — Extract limit/offset from request params,
  use `result.TotalCount` for `TotalItems`. Fixed status validation
  (`unprocessed` → `submitted`).
- **`internal/client/job_list.go`** — Added `limit`/`offset` params to
  `GetJobs()`
- **`internal/client/handler.go`** — Updated `JobHandler` interface
- **`cmd/client_job_list.go`** — Pass limit/offset to API, removed client-side
  slicing, use `TotalItems` from response
- **`cmd/client_job_{get,delete,list}.go`,
  `cmd/client_network_{dns_get,dns_update,ping}.go`,
  `cmd/client_system_{hostname_get,status_get}.go`** — Added missing
  `StatusBadRequest` (400) error handling across all CLI commands
- All tests updated, mocks regenerated, all passing
