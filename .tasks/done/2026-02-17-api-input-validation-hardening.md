---
title: Complete API input validation hardening
status: done
created: 2026-02-17
updated: 2026-02-17
---

## Objective

Harden all API endpoints by validating path parameters, query parameters,
and adding missing 400 responses to OpenAPI specs. The prior validation audit
only covered request body fields via `validate:` struct tags.

## Changes

### OpenAPI specs (400 responses added)

- `internal/api/job/gen/api.yaml` — Added 400 to GET /job, GET /job/{id},
  DELETE /job/{id}
- `internal/api/system/gen/api.yaml` — Added 400 to GET /system/hostname,
  GET /system/status

### Handler validation (8 handlers)

- `network_dns_get_by_interface.go` — interfaceName (required,alphanum) +
  target_hostname (min=1 when provided)
- `network_dns_put_by_interface.go` — target_hostname validation
- `network_ping_post.go` — target_hostname validation
- `system_hostname_get.go` — target_hostname validation
- `system_status_get.go` — target_hostname validation
- `job_get.go` — ID (required,uuid) validation
- `job_delete.go` — ID (required,uuid) validation
- `job_list.go` — status enum validation

### Tests (10 files)

- Updated 5 existing public test files with validation cases
- Created 2 new public test files (system_hostname_get, system_status_get)
- Created 1 new integration test (network_dns_get_by_interface)

## Outcome

All handlers now validate path params, query params, and body fields.
0 lint issues, all tests pass.
