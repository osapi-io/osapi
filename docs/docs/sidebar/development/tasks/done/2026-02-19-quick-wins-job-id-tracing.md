---
title: 'Surface job IDs in API responses and CLI output'
status: done
created: 2026-02-19
updated: 2026-02-20
---

## Objective

CLI commands that internally create NATS jobs (system hostname, system status,
network ping, network DNS) generate a UUID job ID, but that ID is discarded
before reaching the API response or CLI output. Users have no easy way to run
`osapi client job get --job-id <id>` to inspect the job without hunting through
server logs.

Surface the job ID in API responses and CLI output so users can see it and use
it immediately.

## Approach

1. `publishAndWait` / `publishAndCollect` return job ID as first value
2. All `Query*` / `Modify*` methods on `JobClient` interface add `string` (job
   ID) as first return value
3. Regenerate mocks from updated interface
4. Add `job_id` (format: uuid) field to OpenAPI response schemas and regenerate
5. API handlers pass job ID through to the response struct
6. CLI commands display job ID via `printKV` with consistent spacing
7. Rename `RequestID`/`request_id` → `JobID`/`job_id` in job domain types,
   client, and worker code
8. Update CLI docs to show new output format with job IDs
9. Update all tests for new signatures

## Outcome

Job ID is now surfaced in every API response and CLI output. Users can see the
UUID immediately and use `osapi client job get --job-id <id>` to inspect job
status. Naming is consistent (`job_id` everywhere). All tests pass, lint clean,
docs updated.
