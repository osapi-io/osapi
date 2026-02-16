---
title: "Phase 5: Delete legacy task system"
status: done
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Clean break - remove all task code. Job CLI already provides full
replacements for every task CLI command.

## Delete

- `internal/api/task/` (entire directory)
- `internal/task/` (entire directory)
- All `cmd/task*.go` and `cmd/client_task*.go` files
- `test/cli_client_task_integration_test.bats`

## Modify

- `cmd/client_job_delete.go`: Use `jobClient.DeleteJob()`
- `cmd/root.go`: Remove task subcommand if registered
- `internal/config/types.go`: Remove Task config section
- `go.mod`: `go mod tidy`

## Documentation

- Delete `docs/docs/sidebar/usage/cli/task/` (task server/worker docs)
- Delete `docs/docs/sidebar/usage/cli/client/task/` (client task docs)
- Add `docs/docs/sidebar/usage/cli/client/job/` with job CLI docs
  (add, list, get, status, delete, run)
- Update any sidebar config referencing task commands

## Verify

```bash
go build ./...
just test
grep -r "internal/task" --include="*.go" .
```

## Outcome

Completed. All legacy task code removed:
- Deleted `internal/task/` and `internal/api/task/` directories
- Deleted 11 CLI command files (task server/worker/client commands)
- Deleted 5 internal/client task wrapper files
- Removed TaskHandler interface from `internal/client/handler.go`
- Removed Task/TaskServer config structs from `internal/config/types.go`
- Updated `cmd/client_job_delete.go` to use `jobClient.DeleteJob()` instead of raw KV
- Simplified `handler.go`, `manager.go`, `api_server_start.go` (removed task imports/params)
- `go mod tidy` cleaned up unused dependencies
- All tests pass, no `internal/task` references remain in Go files
- Documentation cleanup tracked separately in docs update task
