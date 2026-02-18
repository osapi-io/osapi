---
title: Achieve 100% test coverage on non-generated packages
status: done
created: 2026-02-17
updated: 2026-02-18
---

## Objective

Two non-generated packages are below 100% statement coverage. Add tests for
the uncovered branches to reach full coverage.

## Packages and Gaps

### 1. `internal/api/system` — 98.7% (target: 100%)

**Function:** `GetSystemHostname` in `system_hostname_get.go` — 94.1%

**Uncovered branch (lines 62-65):**

```go
displayHostname := result
if displayHostname == "" {
    displayHostname = workerHostname
}
```

When `QuerySystemHostname` returns an empty string for the display hostname,
the handler falls back to `workerHostname`. No test currently exercises this
path.

**Test to add** in `system_hostname_get_public_test.go`:

- Mock `QuerySystemHostname` to return `("", "worker1", nil)` — empty result
  string with a valid worker hostname.
- Assert response contains `{"results":[{"hostname":"worker1"}]}`.

### 2. `internal/job/client` — 99.6% (target: 100%)

**Function:** `publishAndCollect` in `client.go` — 95.7%

**Uncovered branch A (lines 261-266):** Unmarshal error in broadcast response.

```go
if err := json.Unmarshal(entry.Value(), &response); err != nil {
    c.logger.Warn("failed to unmarshal broadcast response", ...)
    continue
}
```

When a KV watcher entry contains invalid JSON, the response is skipped with a
warning log. No test exercises this.

**Test to add:** Write invalid JSON to the response KV key for a broadcast
job, then write a valid response. Assert only the valid response is collected
and the invalid one is silently skipped.

**Uncovered branch B (lines 270-272):** Empty hostname fallback.

```go
hostname := response.Hostname
if hostname == "" {
    hostname = "unknown"
}
```

When a worker response has an empty `Hostname` field, it is keyed as
`"unknown"` in the results map. No test exercises this.

**Test to add:** Write a valid response to KV with an empty `Hostname` field.
Assert the collected map has a key `"unknown"`.

## Verification

```bash
go test -coverprofile=/tmp/cov.out ./internal/api/system/
go tool cover -func=/tmp/cov.out | grep -v 100.0%
# Should show only total line

go test -coverprofile=/tmp/cov.out ./internal/job/client/
go tool cover -func=/tmp/cov.out | grep -v 100.0%
# Should show only total line
```

## Notes

- Generated packages (`gen/`, `mocks/`) and `cmd/` are excluded from this
  goal — only hand-written business logic packages are targeted.
- The `internal/api/network` package is already at 100%.

## Outcome

All five hand-written packages reached 100% statement coverage as part of the
label-based routing work (commit `e813549`):

- `internal/job` — 100% (added ParseSubject invalid 4-part subject test)
- `internal/job/client` — 100% (added ListWorkers skip tests, refactored
  publishAndCollect to share timeout return path)
- `internal/job/worker` — 100% (added label consumer tests, hostname-with-labels
  processor test)
- `internal/api/system` — 100% (added empty hostname fallback test)
- `internal/api/network` — 100% (already at 100%)
