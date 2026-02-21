---
title: Show worker labels in system hostname output
status: done
created: 2026-02-21
updated: 2026-02-21
---

## Objective

When running `osapi client system hostname` (especially with
`--target _all`), include the labels each worker belongs to. This
gives operators visibility into how their fleet is organized without
having to cross-reference worker configs.

## Current Behavior

```
┏━━━━━━━━━━┓
┃ HOSTNAME ┃
┣━━━━━━━━━━┫
┃ server1  ┃
┃ server2  ┃
┗━━━━━━━━━━┛
```

## Desired Behavior

```
┏━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━┓
┃ HOSTNAME ┃ LABELS              ┃
┣━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━┫
┃ server1  ┃ group:web.dev       ┃
┃ server2  ┃ group:db.prod       ┃
┗━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━┛
```

## Notes

- The worker already sends labels in the `hostname.get` response
  (see `ListWorkers` in `internal/job/client/query.go` which parses
  `labels` from the response). The hostname query handler may just
  need to include labels in its response data.
- Check `internal/job/worker/processor.go` to see if the hostname
  processor already returns labels or if it needs to be added.
- The `job list workers` command already shows labels — reuse that
  pattern.
- For single-host output, labels could be shown as a key-value line
  below hostname rather than a table column.

## Outcome

Implemented across the full stack. Labels now flow from worker
responses through the API and are displayed in CLI output:

- **OpenAPI spec** (`internal/api/system/gen/api.yaml`): Added
  optional `labels` field to `SystemHostnameResult` schema.
- **Job client** (`internal/job/client/query.go`, `types.go`):
  `WorkerInfo` struct carries hostname and labels; returned from
  `QuerySystemHostname` and `QuerySystemHostnameBroadcast`.
- **API handler** (`internal/api/system/system_hostname_get.go`):
  Maps `WorkerInfo.Labels` into the response.
- **CLI** (`cmd/client_system_hostname_get.go`, `cmd/ui.go`):
  Single-host shows `Labels: key:val` via `printKV`; broadcast
  table adds a LABELS column via `formatLabels` helper.
- **Tests**: All four test files audited and confirmed to follow
  repo conventions (suite, table-driven, domain-appropriate
  assertion patterns).
