---
title: Investigate wrapping SDK gen response types
status: backlog
created: 2026-02-24
updated: 2026-02-24
---

## Objective

Investigate whether the SDK should wrap the generated `gen.*Response` types in
SDK-owned types rather than exposing them directly.

Currently, consumers of the SDK (including the osapi CLI) import
`github.com/osapi-io/osapi-sdk/pkg/osapi/gen` to access response types. This
couples consumers to the oapi-codegen output format.

## Considerations

- Wrapping types adds a translation layer but provides stability across codegen
  changes.
- Direct gen types are simpler and avoid duplication, but any codegen change
  (field renames, type changes) ripples to all consumers.
- The CLI currently accesses `resp.JSON200`, `resp.StatusCode()`, `resp.Body`,
  etc. directly on gen response types.
- The `internal/cli/ui.go` and `internal/audit/export/` packages also depend on
  gen types (`gen.JobDetailResponse`, `gen.AuditEntry`, `gen.StatusResponse`,
  `gen.QueueStatsResponse`).

## Notes

This task was created as part of the internal/client to osapi-sdk migration. The
current approach (returning gen types directly) was chosen for simplicity.
Revisit once the SDK API stabilizes.
