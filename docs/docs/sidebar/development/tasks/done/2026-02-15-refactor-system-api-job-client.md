---
title: Port system API to use job client
status: done
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Replace direct provider calls in system API handlers with JobClient calls so
requests route through NATS. System API already uses strict-server.

## Changes

- `internal/api/system/types.go`: Replace providers with JobClient
- `internal/api/system/system.go`: Update factory
- `internal/api/system/system_hostname_get.go`: Use JobClient
- `internal/api/system/system_status_get.go`: Use JobClient
- `internal/api/handler_system.go`: Remove provider construction
- `internal/api/handler.go`: Add jobClient parameter
- `internal/api/manager.go`: Update interface
- `cmd/api_server_start.go`: Wire up job client

## Tests

- `system_hostname_get_test.go` (new internal test)
- `system_status_get_public_test.go` (new)

## Verify

```bash
go build ./...
go test -v ./internal/api/system/...
go test -cover ./internal/api/system/...
```
