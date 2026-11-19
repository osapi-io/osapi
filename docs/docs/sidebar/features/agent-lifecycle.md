---
sidebar_position: 4
---

# Agent lifecycle

osapi agents report **node conditions** that compare values with limits. You can
also **drain and cordon** an agent cleanly for maintenance. The two features
come from the node management of Kubernetes.

## Node conditions

A condition is a boolean that compares a value with a limit. The agent
calculates the conditions at each heartbeat (10 seconds). They show if something
is wrong. The operator does not need to read the raw metrics.

| Condition        | Default Threshold    | Data Source      |
| ---------------- | -------------------- | ---------------- |
| `MemoryPressure` | Memory used > 90%    | Heartbeat memory |
| `HighLoad`       | Load1 > 2x CPU count | Heartbeat load   |
| `DiskPressure`   | Any disk > 90% used  | Heartbeat disk   |

The agent evaluates each condition at each heartbeat, and keeps:

- **Reason**: an explanation for a person, for example "memory 94% used,
  15.1/16.0 GB"
- **LastTransitionTime**: the last time that the condition started or stopped

The API and the CLI show only the conditions that hold now. A condition that
does not hold is not in the answer. The notifier uses the full list of the
agent, so it also sees when a condition stops.

### CLI display

`agent list` shows the conditions that hold in the CONDITIONS column:

```
HOSTNAME  MACHINE ID  STATE  CONDITIONS                LABELS  AGE    LOAD            OS            VERSION
web-01    m-1         Ready  HighLoad, MemoryPressure          3d 4h  4.12 3.90 3.71  Ubuntu 24.04  v1.4.0
web-02    m-2         Ready                                    12h    0.31 0.28 0.25  Ubuntu 24.04  v1.4.0
db-01     m-3         Ready  DiskPressure                      5d     1.22 1.10 1.05  Ubuntu 24.04  v1.3.2
```

The VERSION column is the build of each agent. During an upgrade, it shows which
agents run the new build. The STATE column is the scheduling state: `Ready`,
`Draining`, `Cordoned` or `Pending`.

`agent get` shows each condition that holds, with its reason:

```
Conditions (2):
  TYPE            REASON                                SINCE
  MemoryPressure  memory 94% used (15.1/16.0 GB)        2m
  HighLoad        load 4.12, threshold 4.00 for 2 CPUs  5m
```

### Configuration

Set the limits in `osapi.yaml`:

```yaml
agent:
  conditions:
    memory_pressure_threshold: 90 # percent used
    high_load_multiplier: 2.0 # load1 / cpu_count
    disk_pressure_threshold: 90 # percent used
```

## Agent drain

A drain removes an agent cleanly from the agents that get jobs, for maintenance.
The process continues to run. If an agent stops without a drain, it goes from
the registry, and that looks the same as a crash.

### State machine

Agents have an explicit scheduling state with three values:

```
Ready ──(drain)──> Draining ──(jobs done)──> Cordoned
  ^                                              │
  └──────────────(undrain)───────────────────────┘
```

| State      | Meaning                                         |
| ---------- | ----------------------------------------------- |
| `Ready`    | Accepts and processes jobs (default)            |
| `Draining` | Completes its running jobs, accepts no new jobs |
| `Cordoned` | Fully drained and idle, accepts no jobs         |

### How it works

1. The operator runs `osapi client agent drain --hostname web-01`.
1. The API writes a `drain.{hostname}` key to the state KV bucket.
1. The agent finds the drain flag at its next heartbeat (10s).
1. The agent changes to `Draining` and **stops its NATS JetStream
   subscriptions**. Thus it gets no new jobs.
1. The running jobs continue until they are complete.
1. When all running jobs are complete, the state changes to `Cordoned`.
1. The operator runs `osapi client agent undrain --hostname web-01`.
1. The API deletes the drain key. The agent subscribes again and changes to
   `Ready`.

### Timeline

The agent records each change of state as an event in the state KV bucket
(`agent-state`, no TTL). It only adds events, and does not change them.
`agent get` shows all the changes:

```
Timeline:
  TIMESTAMP              EVENT      HOSTNAME  MESSAGE
  2026-03-05 10:00:00    drain      web-01    Drain initiated
  2026-03-05 10:05:23    cordoned   web-01    All jobs completed
  2026-03-05 12:00:00    undrain    web-01    Resumed accepting jobs
```

### CLI commands

```bash
osapi client agent drain --hostname web-01     # start draining
osapi client agent undrain --hostname web-01   # resume accepting jobs
```

The two commands return the current state and a confirmation message.

## Permissions

The standard `agent:read` responses have the node conditions. Drain and undrain
need the `agent:write` permission. The `admin` role has it by default.

## Related

- [Agent CLI Reference](../usage/cli/client/agent/agent.mdx): agent fleet
  commands
- [Node Management](node-management.md): node queries via the job system
- [Job System](job-system.md): how async job processing works
- [Configuration](../usage/configuration.md): full configuration reference
