---
sidebar_position: 4
---

# Agent Lifecycle

OSAPI agents report threshold-based **node conditions** and support graceful
**drain/cordon** for maintenance. Both features are inspired by Kubernetes node
management patterns.

## Node Conditions

Conditions are threshold-based booleans evaluated agent-side on every heartbeat
(10 seconds). They surface "is anything wrong?" at a glance without requiring
operators to interpret raw metrics.

| Condition        | Default Threshold        | Data Source         |
| ---------------- | ------------------------ | ------------------- |
| `MemoryPressure` | Memory used > 90%        | Heartbeat memory    |
| `HighLoad`       | Load1 > 2x CPU count     | Heartbeat load      |
| `DiskPressure`   | Any disk > 90% used      | Heartbeat disk      |

Each condition tracks:

- **Status** -- `true` when the threshold is exceeded, `false` otherwise
- **Reason** -- human-readable explanation (e.g., "memory 94% used, 15.1/16.0
  GB")
- **LastTransitionTime** -- when the condition last flipped between true and
  false

### CLI Display

`agent list` shows active conditions in the CONDITIONS column:

```
HOSTNAME  STATUS  CONDITIONS               LABELS  AGE    LOAD (1m)  OS
web-01    Ready   HighLoad,MemoryPressure  -       3d 4h  4.12       Ubuntu 24.04
web-02    Ready   -                        -       12h    0.31       Ubuntu 24.04
db-01     Ready   DiskPressure             -       5d     1.22       Ubuntu 24.04
```

`agent get` shows full condition details:

```
Conditions:
  TYPE              STATUS  REASON                                     SINCE
  MemoryPressure    true    memory 94% used (15.1/16.0 GB)             2m ago
  HighLoad          true    load 4.12, threshold 4.00 for 2 CPUs       5m ago
  DiskPressure      false
```

### Configuration

Thresholds are configurable in `osapi.yaml`:

```yaml
agent:
  conditions:
    memory_pressure_threshold: 90   # percent used
    high_load_multiplier: 2.0       # load1 / cpu_count
    disk_pressure_threshold: 90     # percent used
```

## Agent Drain

Drain allows operators to gracefully remove an agent from the job routing pool
for maintenance without stopping the process. When an agent stops without
draining, it vanishes from the registry and looks identical to a crash.

### State Machine

Agents have an explicit scheduling state with three values:

```
Ready ──(drain)──> Draining ──(jobs done)──> Cordoned
  ^                                              │
  └──────────────(undrain)───────────────────────┘
```

| State      | Meaning                                      |
| ---------- | -------------------------------------------- |
| `Ready`    | Accepting and processing jobs (default)      |
| `Draining` | Finishing in-flight jobs, not accepting new   |
| `Cordoned` | Fully drained, idle, not accepting jobs       |

### How It Works

1. Operator calls `osapi client agent drain --hostname web-01`
2. API writes a `drain.{hostname}` key to the registry KV bucket
3. Agent detects the drain flag on its next heartbeat tick (10s)
4. Agent transitions to `Draining` and **unsubscribes from NATS JetStream
   consumers** -- this is how it stops receiving new jobs
5. In-flight jobs continue to completion
6. Once all in-flight jobs finish, state becomes `Cordoned`
7. Operator calls `osapi client agent undrain --hostname web-01`
8. API deletes the drain key; agent resubscribes and transitions to `Ready`

### Timeline

Every state transition is recorded as an append-only event in the registry KV
bucket. `agent get` shows the full transition history:

```
Timeline:
  TIMESTAMP              EVENT      HOSTNAME  MESSAGE
  2026-03-05 10:00:00    drain      web-01    Drain initiated
  2026-03-05 10:05:23    cordoned   web-01    All jobs completed
  2026-03-05 12:00:00    undrain    web-01    Resumed accepting jobs
```

### CLI Commands

```bash
osapi client agent drain --hostname web-01     # start draining
osapi client agent undrain --hostname web-01   # resume accepting jobs
```

Both commands return the current state and a confirmation message.

## Permissions

Node conditions are included in the standard `agent:read` responses. Drain and
undrain operations require the `agent:write` permission, which is included in
the `admin` role by default.

## Related

- [Agent CLI Reference](../usage/cli/client/agent/agent.mdx) -- agent fleet
  commands
- [Node Management](node-management.md) -- node queries via the job system
- [Job System](job-system.md) -- how async job processing works
- [Configuration](../usage/configuration.md) -- full configuration reference
