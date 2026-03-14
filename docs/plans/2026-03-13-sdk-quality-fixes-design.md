# SDK Quality Fixes Design

## Problem

A code review identified 9 issues across the OSAPI SDK and its primary
consumer (`osapi-orchestrator`). The central gap is that the SDK's bridge
helpers are incomplete, forcing `osapi-orchestrator` to reimplement result
conversion and duplicate ~200 lines of type definitions.

## Fixes — osapi SDK (this repo)

### #1: CollectionResult populates Result.Data

`CollectionResult` currently only populates `HostResult.Data` per-host.
It leaves `Result.Data` nil, so `osapi-orchestrator` must call
`mustRawToMap(resp.RawJSON())` separately for every operation.

**Fix:** Add a `rawJSON []byte` parameter. When non-nil, unmarshal it
into `Result.Data`. Callers pass `resp.RawJSON()` or `nil`.

```go
func CollectionResult[T any](
    col Collection[T],
    rawJSON []byte,
    toHostResult func(T) HostResult,
) *Result
```

Update all callers in examples and tests.

### #2: Docker SDK request types

`DockerService.Create`, `List`, `Stop`, and `Remove` expose `gen.*`
request types. Every other service wraps gen types into SDK-defined
types.

**Fix:** Define in `docker_types.go`:

- `DockerCreateOpts` — Image, Name, Command, Env, Ports, Volumes,
  AutoStart
- `DockerStopOpts` — Timeout
- `DockerListParams` — State
- `DockerRemoveParams` — Force

Map to gen types inside the service methods. Consumers no longer
import `gen`.

### #5: Collection[T].First()

Every consumer blindly indexes `Results[0]` with no bounds check.

**Fix:** Add to `Collection[T]` in `response.go`:

```go
func (c Collection[T]) First() (T, bool) {
    if len(c.Results) == 0 {
        var zero T
        return zero, false
    }
    return c.Results[0], true
}
```

### #7: JSON tags on SDK result types

SDK result types (`HostnameResult`, `DiskResult`, `CommandResult`,
etc.) lack `json:"..."` tags. `StructToMap` cannot produce correct
keys without them, forcing consumers to use `RawJSON()` as a
workaround.

**Fix:** Add `json` tags to all result types in `node_types.go`,
`docker_types.go`, `file_types.go`, `audit_types.go`,
`health_types.go`, `job_types.go`, `agent_types.go`.

### #8: AuditService.Get UUID error wrapping

`AuditService.Get` returns raw UUID parse error without context.
`JobService.Get` wraps it correctly.

**Fix:** Wrap with `fmt.Errorf("invalid audit ID: %w", err)`.

## Fixes — osapi-orchestrator (separate repo)

### #4: mustRawToMap panic → error return

`mustRawToMap` panics on invalid JSON. A proxy 502 or truncated
response would crash the process.

**Fix:** Change return to `(map[string]any, error)`. Propagate
error through all callers.

After SDK fix #1 lands, `mustRawToMap` can be deleted entirely
since `CollectionResult` handles raw JSON internally.

### #6: Delete duplicated types

`result_types.go` (~200 lines) redefines SDK types: `HostnameResult`,
`DiskResult`, `MemoryResult`, `LoadResult`, `CommandResult`,
`PingResult`, `DNSConfigResult`, `DNSUpdateResult`, `FileDeployOpts`,
`FileDeployResult`, `FileStatusResult`, `FileUploadResult`,
`FileChangedResult`, `AgentResult`, `AgentListResult`, plus sub-types.

**Fix:** Delete `result_types.go`. Use `client.*` types directly
throughout `ops.go` and any other files that reference these types.
This eliminates the duplicate definitions and the field-by-field
copies in `ops.go`.

### #9: HealthCheck target parameter

`HealthCheck` accepts a `target` parameter but ignores it. Liveness
checks hit the API server directly — target routing doesn't apply.

**Fix:** Remove the unused parameter.

### #10: Report.Summary() duplication

The orchestrator reimplements `Summary()` instead of delegating
to the SDK's version.

**Fix:** Delegate to `sdk.Report.Summary()`.

### Additional: Replace buildResult/toMap with SDK helpers

Once SDK fixes #1 and #7 land:

- Replace local `buildResult` with `orchestrator.CollectionResult`
- Replace local `toMap` with `orchestrator.StructToMap`
- Delete `mustRawToMap` (no longer needed)

## What Does NOT Change

- The orchestrator DAG engine (plan, task, runner) — already solid
- `Response[T]` and `Collection[T]` pattern — correct design
- Error hierarchy (`checkError`, `AuthError`, etc.) — clean
- `MetricsService` using `http.DefaultClient` — intentional, `/metrics`
  is a Prometheus endpoint outside the auth middleware

## Order of Operations

1. Fix osapi SDK first (#1, #2, #5, #7, #8) — with tests, 100%
   coverage
2. Fix osapi-orchestrator (#4, #6, #9, #10, plus adopt SDK helpers)
   — depends on SDK changes being published
