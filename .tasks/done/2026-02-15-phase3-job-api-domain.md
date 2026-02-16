---
title: "Phase 3: Create job API domain (replace task API)"
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

New `internal/api/job/` domain with strict-server providing REST endpoints
for job management. Replaces legacy `internal/api/task/`.

## Changes

- New OpenAPI spec, codegen config, and generated code
- New handler files for each endpoint (create, list, get, status, delete)
- Add `DeleteJob` to JobClient interface + implementation
- Wire into `internal/api/handler.go`

## Endpoints

- `POST /job` - Create job
- `GET /job` - List jobs
- `GET /job/status` - Queue statistics
- `GET /job/{id}` - Get job detail
- `DELETE /job/{id}` - Delete job

## Tests

- Public test file for each handler
- `DeleteJob` client test

## Verify

```bash
go generate ./internal/api/job/gen/...
go generate ./internal/job/mocks/...
go build ./...
go test -v ./internal/api/job/...
```
