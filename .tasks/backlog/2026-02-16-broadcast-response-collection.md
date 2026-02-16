---
title: Implement broadcast (_all) response collection
status: backlog
created: 2026-02-16
updated: 2026-02-16
---

## Objective

The multi-system routing verification (2026-02-15) confirmed that `_all`
broadcast routing works correctly at the NATS level, but three gaps exist
in the response collection layer:

1. **`QuerySystemStatusAll` is a stub** — The job client method for
   querying all systems' status is not yet implemented.
2. **`publishAndWait` collects only the first response** — For `_all`
   jobs, the client needs to wait for and aggregate responses from
   multiple workers.
3. **REST API single-result schema** — `GET /job/{id}` returns a single
   `hostname`/`result` pair. For `_all` jobs, the response should include
   per-worker results (list of hostname/result/status tuples).

## Tasks

- [ ] Implement `QuerySystemStatusAll` in `internal/job/client/query.go`
- [ ] Extend `publishAndWait` (or create `publishAndCollect`) to gather
      multiple responses within a timeout window
- [ ] Add `workers` field (or similar) to `JobDetailResponse` schema to
      return per-worker results for broadcast jobs
- [ ] Add tests for multi-worker response aggregation

## Notes

- The append-only KV architecture already stores per-worker responses at
  `responses.{job-id}.{hostname}.{nano}` — the data is there, we just
  need to read and aggregate it.
- Consider a configurable timeout for how long to wait for all workers to
  respond before returning partial results.
