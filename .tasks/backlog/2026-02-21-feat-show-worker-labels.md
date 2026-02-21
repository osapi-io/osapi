---
title: Show worker labels in system hostname output
status: backlog
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
