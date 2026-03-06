# Node Conditions and Agent Drain

## Context

OSAPI agents collect rich system metrics (memory, load, disk, CPU count) via
heartbeat and facts, but operators must manually interpret raw numbers to detect
problems. Kubernetes solves this with node conditions — threshold-based booleans
that surface "is anything wrong?" at a glance.

Additionally, there's no way to gracefully remove an agent from the job routing
pool for maintenance without stopping the process entirely. When an agent stops,
it vanishes from the registry and looks identical to a crash. Kubernetes handles
this with cordon/drain.

This design adds both features to OSAPI.

## Node Conditions

### Condition Types

Three conditions derived from existing heartbeat and facts data, evaluated
agent-side on each heartbeat tick (10s):

| Condition        | Default Threshold    | Data Source                                     |
| ---------------- | -------------------- | ----------------------------------------------- |
| `MemoryPressure` | memory used > 90%    | `MemoryStats` (heartbeat)                       |
| `HighLoad`       | load1 > 2× CPU count | `LoadAverages` (heartbeat) + `CPUCount` (facts) |
| `DiskPressure`   | any disk > 90% used  | `DiskStats` (new in heartbeat)                  |

### Condition Structure

Each condition has:

```go
type Condition struct {
    Type               string    `json:"type"`
    Status             bool      `json:"status"`
    Reason             string    `json:"reason,omitempty"`
    LastTransitionTime time.Time `json:"last_transition_time"`
}
```

- `Status`: `true` = condition is active (pressure/overload detected)
- `Reason`: human-readable explanation (e.g., "memory 94% used (15.1/16.0 GB)")
- `LastTransitionTime`: when the condition last changed from true→false or
  false→true

### Configuration

Thresholds configurable in `osapi.yaml` with sensible defaults:

```yaml
agent:
  conditions:
    memory_pressure_threshold: 90 # percent used
    high_load_multiplier: 2.0 # load1 / cpu_count
    disk_pressure_threshold: 90 # percent used
```

### Evaluation

Conditions are evaluated in the agent during `writeRegistration()`. The agent
maintains previous condition state in memory to track `LastTransitionTime` —
only updated when the boolean flips.

DiskPressure requires adding disk stats to the heartbeat. The existing
`disk.Provider` already implements `GetUsage()` so the data is available. Disk
collection joins the existing non-fatal provider pattern: if it fails, the
DiskPressure condition is simply not evaluated.

### Storage

Conditions are stored as part of `AgentRegistration` in the registry KV bucket.
No new KV bucket needed.

```go
type AgentRegistration struct {
    // ... existing fields ...
    Conditions []Condition `json:"conditions,omitempty"`
}
```

### CLI Display

`agent list` gains a CONDITIONS column showing active conditions:

```
HOSTNAME    STATUS   CONDITIONS              LOAD   OS
web-01      Ready    HighLoad,MemoryPressure  4.12   Ubuntu 24.04
web-02      Ready    -                        0.31   Ubuntu 24.04
db-01       Ready    DiskPressure             1.22   Ubuntu 24.04
```

`agent get` shows full condition details and state timeline:

```
Conditions:
  MemoryPressure: true  (memory 94% used, 15.1/16.0 GB)  since 2m ago
  HighLoad: true  (load 4.12, threshold 4.00 for 2 CPUs)  since 5m ago
  DiskPressure: false

Timeline:
  TIMESTAMP              EVENT      HOSTNAME    MESSAGE
  2026-03-05 10:00:00    drain      web-01      Drain initiated
  2026-03-05 10:05:23    cordoned   web-01      All jobs completed
  2026-03-05 12:00:00    undrain    web-01      Resumed accepting jobs
```

## Agent Drain

### State Machine

Agents gain an explicit state field with three values:

```
Ready ──(drain)──> Draining ──(jobs done)──> Cordoned
  ^                                              │
  └──────────────(undrain)───────────────────────┘
```

| State      | Meaning                                          |
| ---------- | ------------------------------------------------ |
| `Ready`    | Accepting and processing jobs (default)          |
| `Draining` | Finishing in-flight jobs, not accepting new ones |
| `Cordoned` | Fully drained, idle, not accepting jobs          |

### Mechanism

1. Operator calls `POST /agent/{hostname}/drain`
2. API writes a `drain.{hostname}` key to the state KV bucket
3. Agent checks for drain key on each heartbeat tick (10s)
4. When drain flag detected:
   - Agent transitions state to `Draining`
   - Agent unsubscribes from NATS consumer (stops receiving new jobs)
   - In-flight jobs continue to completion
5. Once WaitGroup drains (no in-flight jobs), state becomes `Cordoned`
6. `POST /agent/{hostname}/undrain` deletes the drain key
7. Agent detects drain key removal on next heartbeat:
   - Transitions state to `Ready`
   - Re-subscribes to NATS consumer

### API Endpoints

```
POST /agent/{hostname}/drain     # Start draining
POST /agent/{hostname}/undrain   # Resume accepting jobs
```

Both return 200 on success, 404 if agent not found, 409 if already in the
requested state.

### Permission

New `agent:write` permission. Added to the `admin` role by default.

### Storage

Agent state transitions are recorded as **append-only events** in the state KV
bucket (`agent-state`, no TTL), following the same pattern used for job status
events (see `WriteStatusEvent` in `internal/job/client/agent.go`).

Events reuse the existing `TimelineEvent` type (`internal/job/types.go`) — the
same type used for job lifecycle events. This type is generic (Timestamp, Event,
Hostname, Message, Error) and not job-specific:

```
Key format: timeline.{sanitized_hostname}.{event}.{unix_nano}
Value:      TimelineEvent JSON
```

Events: `ready`, `drain`, `cordoned`, `undrain`

On the SDK side, `TimelineEvent` is promoted from `job_types.go` to a shared
top-level type in `pkg/osapi/types.go`. Both `JobDetail.Timeline` and
`Agent.Timeline` reference the same type.

Current state is **computed from the latest event**, just like job status is
computed via `computeStatusFromEvents`. This preserves the full transition
history (Ready → Draining → Cordoned → Ready → Draining → ...) and eliminates
race conditions by never updating existing keys.

The drain intent uses a separate key: `drain.{sanitized_hostname}`. The API
writes this key to signal drain; the agent reads it on heartbeat and writes the
state transition event. The API deletes the key on undrain.

The `AgentRegistration` also carries the current state for quick reads without
scanning events:

```go
type AgentRegistration struct {
    // ... existing fields ...
    State string `json:"state,omitempty"` // Ready, Draining, Cordoned
}
```

### CLI Commands

```bash
osapi client agent drain --hostname web-01
osapi client agent undrain --hostname web-01
```

`agent list` and `agent get` show the state in the STATUS column.

## OpenAPI Changes

### AgentInfo Schema

Add to existing `AgentInfo`:

```yaml
state:
  type: string
  enum: [Ready, Draining, Cordoned]
  description: Agent scheduling state.
conditions:
  type: array
  items:
    $ref: '#/components/schemas/NodeCondition'
```

New schema:

```yaml
NodeCondition:
  type: object
  properties:
    type:
      type: string
      enum: [MemoryPressure, HighLoad, DiskPressure]
    status:
      type: boolean
    reason:
      type: string
    last_transition_time:
      type: string
      format: date-time
  required: [type, status, last_transition_time]
```

### New Endpoints

```yaml
/agent/{hostname}/drain:
  post:
    summary: Drain an agent
    description: Stop the agent from accepting new jobs.
    security:
      - BearerAuth: []
    responses:
      200: ...
      404: ...
      409: ...

/agent/{hostname}/undrain:
  post:
    summary: Undrain an agent
    description: Resume accepting jobs on a drained agent.
    security:
      - BearerAuth: []
    responses:
      200: ...
      404: ...
      409: ...
```

### Permission Updates

```yaml
# New permission
agent:write

# Updated admin role
admin:
  permissions:
    - agent:read
    - agent:write    # new
    - node:read
    - ...
```

## Implementation Scope

### Provider Changes

- Extend heartbeat to collect disk stats (reuse existing `disk.Provider`)
- Add condition evaluation logic to agent heartbeat

### Agent Changes

- Add `Condition` type and evaluation functions
- Add state field to `AgentRegistration`
- Add drain flag detection on heartbeat tick
- Add consumer subscribe/unsubscribe for drain/undrain transitions
- Add condition threshold config support

### API Changes

- New drain/undrain endpoints in the agent API domain
- Extend `AgentInfo` schema with `state` and `conditions`
- Add `agent:write` permission and wire into scope middleware

### CLI Changes

- `agent drain` and `agent undrain` commands
- CONDITIONS column in `agent list`
- Condition details and state timeline in `agent get`
- State shown in STATUS column

### SDK Changes

- Promote `TimelineEvent` from `job_types.go` to shared `types.go`
- Both `JobDetail.Timeline` and `Agent.Timeline` use the same type
- Add `Agent.Drain()` and `Agent.Undrain()` methods
- Add conditions, state, and timeline to `Agent` type

### Config Changes

- `agent.conditions` section with threshold defaults

## Testing

- **Unit**: condition evaluation logic (threshold math, transition tracking),
  state machine transitions, drain flag detection
- **HTTP wiring**: drain/undrain endpoints with RBAC (401, 403, 200, 404, 409)
- **Integration**: drain agent → submit job → verify not routed to drained agent
  → undrain → verify jobs resume

## Verification

```bash
just generate        # regenerate specs + code
go build ./...       # compiles
just go::unit        # tests pass
just go::vet         # lint passes
```
