---
title: "Phase 1: Fix operation naming inconsistencies"
status: in-progress
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Align operation names across constants, client, and worker. Foundational
phase that all subsequent phases depend on.

## Changes

- `internal/job/types.go`: Change `OperationNetworkPingExecute` from
  `"network.ping.execute"` to `"network.ping.do"`
- `internal/job/client/modify.go`: Fix `Operation: "dns.set"` to
  `"dns.update"`, remove ping methods
- `internal/job/client/query.go`: Add `QueryNetworkPing` and
  `QueryNetworkPingAny` (moved from modify, renamed, use TypeQuery)
- `internal/job/client/types.go`: Update interface with renamed methods
- Regenerate mocks
- Move ping tests from `modify_public_test.go` to `query_public_test.go`

## Verify

```bash
go build ./...
go test -v ./internal/job/...
```

## Notes

- Worker processor uses `"ping"` as category switch case (not the full
  operation string), so worker code doesn't need changes for the rename
- The `"ping.do"` operation string in modify.go already matches the target

## Outcome

_To be filled when complete._
