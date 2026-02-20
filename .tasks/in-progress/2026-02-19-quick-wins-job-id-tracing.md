---
title: "Surface job IDs in API responses and CLI output"
status: in-progress
created: 2026-02-19
updated: 2026-02-19
---

## Objective

CLI commands that internally create NATS jobs (system hostname, system
status, network ping, network DNS) generate a UUID job ID, but that ID
is discarded before reaching the API response or CLI output. Users have
no easy way to run `osapi client job get --job-id <id>` to inspect the
job without hunting through server logs.

Surface the job ID in API responses and CLI output so users can see it
and use it immediately.

**Explicitly out of scope:** Adding `slog` calls in handlers or API
layers for correlation. That is a distributed tracing concern solved
properly by OTel/Jaeger (see backlog task).

## Approach

1. `publishAndWait` / `publishAndCollect` return job ID as first value
2. All `Query*` / `Modify*` methods on `JobClient` interface add
   `string` (job ID) as first return value
3. Regenerate mocks from updated interface
4. Add `job_id` (format: uuid) field to OpenAPI response schemas and
   regenerate
5. API handlers pass job ID through to the response struct (no slog
   additions)
6. CLI commands display job ID via `printKV`
7. Update all tests for new signatures

## Files to Modify

- `internal/job/client/client.go` -- publishAndWait/publishAndCollect
- `internal/job/client/types.go` -- JobClient interface
- `internal/job/client/query.go` -- all Query methods
- `internal/job/client/modify.go` -- all Modify methods
- `internal/job/mocks/job_client.gen.go` -- regenerated
- `internal/api/system/gen/api.yaml` -- add job_id to schemas
- `internal/api/network/gen/api.yaml` -- add job_id to schemas
- `internal/api/system/system_hostname_get.go`
- `internal/api/system/system_status_get.go`
- `internal/api/network/network_ping_post.go`
- `internal/api/network/network_dns_get_by_interface.go`
- `internal/api/network/network_dns_put_by_interface.go`
- `cmd/client_system_hostname_get.go`
- `cmd/client_system_status_get.go`
- `cmd/client_network_ping.go`
- `cmd/client_network_dns_get.go`
- `cmd/client_network_dns_update.go`
- All corresponding `*_public_test.go` and `*_integration_test.go`

## Outcome

_To be filled in when done._
