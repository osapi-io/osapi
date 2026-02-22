---
title: Verify multi-system routing and per-system response aggregation
status: done
created: 2026-02-15
updated: 2026-02-16
---

## Objective

Verify that the job system architecture correctly supports routing messages to a
single system or multiple systems, and that responses are aggregated per-system
so consumers can see what each individual system reported back.

## Key Questions to Verify

1. **Single-system routing (`_any`)**: Does a job targeted at `_any` get picked
   up by exactly one worker and return that worker's response?

2. **Multi-system routing (`_all`)**: Does a job targeted at `_all` get
   broadcast to all workers, and do all workers process independently?

3. **Per-system responses**: When multiple systems process the same job, can the
   client see each system's individual response (hostname, result, status,
   timing)?

4. **Specific-host routing (`hostname`)**: Does a job targeted at a specific
   hostname only get processed by that worker?

5. **Response aggregation**: For `_all` jobs, does the status computation
   correctly show `partial_failure` when some systems succeed and others fail?

## Architecture Reference

The append-only status architecture (`docs/docs/sidebar/architecture.md`)
already describes:

- `status.{job-id}.{event}.{hostname}.{nano}` for per-worker status events
- `responses.{job-id}.{hostname}.{nano}` for per-worker responses
- Computed status from events (completed, partial_failure, failed)

## What to Verify

- Walk through the code path from `CreateJob` through subject routing
  (`jobs.{type}.{hostname}`) to worker subscription patterns
- Verify `_all` broadcast works without queue groups
- Verify `_any` load balancing works with queue groups
- Verify the response KV keys include hostname for per-system attribution
- Verify `GetJobStatus` aggregates multi-worker responses correctly
- Consider whether the REST API `GET /job/{id}` response needs enhancement to
  return a list of per-worker results for `_all` jobs (currently returns single
  hostname/result)

## Acceptance Criteria

- Document any gaps found between the architecture doc and implementation
- If gaps exist, create follow-up tasks to fix them
- Ensure integration tests or manual testing cover single-host, any-host, and
  all-host routing scenarios

## Outcome

Architecture verification confirmed routing is **correct**:

- `_any` → queue groups (load balanced via NATS consumer)
- `_all` → no queue groups (broadcast to all workers)
- Direct hostname → filtered consumer on specific host
- Append-only events with nanosecond timestamps eliminate race conditions
- `partial_failure` status aggregation works correctly

**Gaps identified** (follow-up task created):

- `QuerySystemStatusAll` is a stub — `_all` broadcast response collection is not
  yet implemented
- `publishAndWait` only collects the first response
- REST API `GET /job/{id}` returns a single result, not per-worker results

See: `.tasks/backlog/2026-02-16-broadcast-response-collection.md`
