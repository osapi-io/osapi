# Broadcast Status Reporting Design

## Goal

Stop silently dropping failed/skipped agent responses in broadcast
operations. Every targeted agent should appear in the results — whether
it succeeded, failed, or was skipped — so users, the SDK, and the
orchestrator have full visibility into what happened across the fleet.

## Problem

When targeting `_all` or label selectors, the API handler filters out
failed and skipped responses:

```go
if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
    continue
}
```

This means the CLI shows partial results with no indication that agents
were missing. The SDK and orchestrator never see the failed/skipped
hosts, so guards like `OnlyIfAnyHostFailed` can't detect partial
failures.

## Design

### Approach: Stop Filtering

The fix is narrow: remove the filter at the API handler layer. Pass all
agent responses through the full stack. Every result type already has
`Error` and `Hostname` fields. The CLI already conditionally shows
STATUS/ERROR columns when errors are present. The orchestrator's
`CollectionResult` bridge and host-level guards work unchanged.

No new types, no new endpoints, no schema changes.

### Data Flow (After Fix)

```
Agent 1,2,3,4,5 (broadcast target)
        ↓
publishAndCollect() collects all 5 responses
        ↓
API handler iterates ALL responses
    ├─ Agent 1: success → result entry
    ├─ Agent 2: success → result entry
    ├─ Agent 3: FAILED → result entry (hostname + error)
    ├─ Agent 4: success → result entry
    └─ Agent 5: SKIPPED → result entry (hostname + error)
        ↓
SDK Collection: 5 results (3 ok, 1 failed, 1 skipped)
        ↓
CLI table: all 5 rows, STATUS/ERROR columns visible
        ↓
Orchestrator: 5 HostResults, guards see failed hosts
```

### Changes by Layer

**API handlers** — every broadcast handler has the same pattern:

```go
for _, resp := range responses {
    if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
        continue  // ← REMOVE THIS
    }
    allResults = append(allResults, responseTo*(resp)...)
}
```

Remove the filter. For failed/skipped responses, produce a result entry
with the hostname and error populated, domain-specific fields empty.

Affected handlers (every handler that calls a `*Broadcast` method):
- `schedule/cron_list_get.go` — `getNodeScheduleCronBroadcast`
- `node/node_hostname_get.go` — broadcast path
- `node/node_status_get.go` — broadcast path
- `node/node_*_get.go` — all node query handlers with broadcast
- `node/network_dns_get.go` — broadcast path
- `node/command_exec_post.go` — broadcast path
- `node/command_shell_post.go` — broadcast path
- `docker/container_list.go` — broadcast path
- Any other handler using `IsBroadcastTarget`

**Response conversion** — each handler's `responseTo*()` function
(or inline conversion loop) needs a branch for failed/skipped
responses that produces an entry with:
- `Hostname` from `resp.Hostname`
- `Error` from `resp.Error` (or a descriptive message for skipped)
- All domain fields empty/zero

**SDK types** — no changes. Every result type already has:
- `Hostname string`
- `Error string`

**SDK Collection** — no changes. `Collection[T]` already holds all
results in `Results []T`.

**CLI** — no changes to `BuildBroadcastTable`. It already:
- Shows STATUS column when any result has a non-empty status
- Shows ERROR column when any result has an error
- Displays hostname per row

The output becomes:

```
HOSTNAME  STATUS   NAME         SCHEDULE   OBJECT         USER
web-01    ok       test-backup  0 2 * * *  backup-script  root
web-02    ok       test-backup  0 2 * * *  backup-script  root
web-03    failed   -            -          -              -     connection timeout
web-05    skipped  -            -          -              -     unsupported OS family
```

**Orchestrator** — no changes. `CollectionResult()` already maps each
SDK result to a `HostResult`:

```go
orchestrator.HostResult{
    Hostname: r.Hostname,
    Changed:  r.Changed,
    Error:    r.Error,  // populated for failed/skipped
}
```

Guards like `OnlyIfAnyHostFailed()` check `Error != ""` — they
naturally detect the failed hosts that were previously invisible.

### Edge Cases

- **All agents skipped**: table shows all rows as skipped. No
  successful results but the user sees why.
- **Timeout (agent didn't respond)**: `publishAndCollect` already
  returns only agents that responded within the timeout. Non-responding
  agents are absent. This is a separate problem (requires comparing
  expected vs actual) and is out of scope.
- **Single-target operations**: unchanged. `publishAndWait` returns
  one response, no filtering involved.

### What This Does NOT Change

- Single-target (`_any`, specific hostname) operations — unchanged
- The `publishAndCollect` / `publishAndWait` job client methods
- OpenAPI response schemas — result arrays already support error fields
- SDK type definitions
- CLI table rendering logic
- Orchestrator bridge helpers or guards
