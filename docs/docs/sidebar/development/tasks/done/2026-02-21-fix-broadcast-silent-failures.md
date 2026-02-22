---
title: Surface failed worker responses in broadcast query results
status: done
created: 2026-02-21
updated: 2026-02-21
---

## Objective

When using `--target _all` or label-based targets, workers that fail (e.g.,
interface not found, permission denied) are silently dropped from the response.
The user gets no indication that a worker was reached but failed.

**Example**: `osapi client network dns get --interface eth0 --target _all` only
shows the Mac worker because the Linux worker fails with "interface eth0 does
not exist" — but the user sees no error.

## Current Behavior

All broadcast query methods (`QuerySystemStatusBroadcast`,
`QueryNetworkDNSBroadcast`, `QueryNetworkPingBroadcast`, etc.) in
`internal/job/client/query.go` skip failed responses silently:

```go
if resp.Status == "failed" {
    continue
}
```

## Options

1. **Log skipped failures** (quick win) — when filtering failed responses, log a
   warning like
   `"skipping failed broadcast response" hostname=nerd error="interface eth0 does not exist"`.
   Operators can see what was dropped.

2. **Include errors in API response** — return failed workers in the response
   with an error field, so the CLI can show a row like
   `nerd | ERROR: interface "eth0" does not exist`. This requires OpenAPI schema
   changes.

## Files Involved

| File                           | Role                                                             |
| ------------------------------ | ---------------------------------------------------------------- |
| `internal/job/client/query.go` | All broadcast query methods with `resp.Status == "failed"` skips |

## Notes

- This is inherent to heterogeneous fleets — Linux uses eth0/eno1, macOS uses
  en0. Interface names are required path parameters.
- The "skip failed" behavior is reasonable for broadcast, but users need
  visibility into which workers were dropped and why.
