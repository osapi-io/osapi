---
title: Implement broadcast (_all) response collection
status: done
created: 2026-02-16
updated: 2026-02-17
---

## Objective

The multi-system routing verification (2026-02-15) confirmed that `_all`
broadcast routing works correctly at the NATS level, but three gaps exist in the
response collection layer:

1. **`QuerySystemStatusAll` is a stub** — The job client method for querying all
   systems' status is not yet implemented.
2. **`publishAndWait` collects only the first response** — For `_all` jobs, the
   client needs to wait for and aggregate responses from multiple workers.
3. **REST API single-result schema** — `GET /job/{id}` returns a single
   `hostname`/`result` pair. For `_all` jobs, the response should include
   per-worker results (list of hostname/result/status tuples).

## Tasks

- [x] Implement `QuerySystemStatusAll` in `internal/job/client/query.go`
- [x] Create `publishAndCollect` to gather multiple responses within a timeout
      window
- [x] Add `responses` and `worker_states` fields to `JobDetailResponse` schema
      to return per-worker results for broadcast jobs
- [x] Add tests for multi-worker response aggregation
- [x] Fix `GetJobStatus` hostname/result mapping gap
- [x] Add `QuerySystemStatusAll` to `JobClient` interface

## Outcome

All broadcast response collection gaps resolved:

- `publishAndCollect()` method waits full timeout, collects all worker responses
  into `map[string]*job.Response` keyed by hostname
- `QuerySystemStatusAll()` uses `publishAndCollect` with `_all` subject
- REST API `GET /job/{id}` exposes `responses` (per-worker results) and
  `worker_states` (per-worker state/duration) for broadcast jobs
- `GetJobStatus` now populates `Hostname` from computed status and `Result` from
  single-worker response
- All tests pass, zero lint issues
