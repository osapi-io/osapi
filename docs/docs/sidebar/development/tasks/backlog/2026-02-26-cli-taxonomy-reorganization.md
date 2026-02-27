---
title: CLI taxonomy reorganization
status: backlog
created: 2026-02-26
updated: 2026-02-26
---

## Objective

Rethink the CLI command hierarchy to better reflect operational mental models.
The current structure mixes concerns — workers live under `job`, fleet-wide
system info is under `system`, and `health` is ambiguous between control-plane
health and fleet health. Drawing parallels to Kubernetes (nodes, pods, kubelet),
the CLI should make it obvious where to look when you want to answer: "is my
fleet healthy and what's out there?"

## Current State

```
osapi client
├── health
│   ├── liveness     # API server liveness probe
│   ├── ready        # API server readiness probe
│   └── status       # API server + NATS component health
├── system
│   ├── hostname     # Get hostname (via job broadcast)
│   ├── status       # OS-level info: load, memory, disks, uptime
│   ├── exec         # Run a command on target(s)
│   └── shell        # Run a shell command on target(s)
├── job
│   ├── add / get / list / run / status
│   └── workers
│       └── list     # Discover active workers (hostname only)
├── network
│   ├── dns (get/update)
│   └── ping
├── audit (list/get/export)
└── metrics
```

## Problems

1. **`job workers list`** is the fleet discovery command but it's buried under
   `job`. Workers aren't a sub-concept of jobs — they're the fleet. Like
   `kubectl get nodes` vs `kubectl get jobs`.

2. **Workers only show hostname** — no labels, no status, no uptime. With the
   new heartbeat registry (KV-based), we have richer data available.

3. **`health` is ambiguous** — it shows API server + NATS infrastructure health,
   but at the top level it reads like "is my fleet healthy?" It's really a
   control-plane operational endpoint (like `/healthz` on the Kubernetes API
   server).

4. **`system hostname`** is mostly redundant now that `system status` includes
   hostname and `exec` exists. May have been useful early on but doesn't carry
   its weight.

5. **No single "operational dashboard"** — if you want to know "is everything
   working?", there's no obvious place to start.

## Ideas to Discuss

### Option A: Introduce `node` as a top-level concept

```
osapi client
├── node
│   ├── list         # Fleet view: hostname, labels, status, uptime
│   └── status       # Detailed single-node view (what system status does)
├── system
│   ├── exec
│   └── shell
├── health           # Stays, but clearly scoped as control-plane health
│   ├── liveness
│   ├── ready
│   └── status       # API server + NATS internals
├── job (unchanged)
├── network (unchanged)
├── audit (unchanged)
└── metrics
```

- `node list` becomes `kubectl get nodes` — shows all workers with hostname,
  labels, status (from the heartbeat registry)
- `node status <hostname>` shows the detailed OS view (load, memory, disks) —
  what `system status` does today
- `system` shrinks to just exec/shell (remote execution)
- Workers get a better name ("nodes"? something else?)

### Option B: Rename `job workers` to something better, keep structure

Less disruptive — just promote workers to top-level and add labels:

```
osapi client
├── fleet
│   └── list         # Workers with labels, status
├── system (unchanged)
├── health (unchanged)
...
```

### Option C: Merge health + fleet into a single operational view

```
osapi client status   # One-stop: API health + worker fleet + job summary
```

### Open Questions

- What's the right name? `node`, `agent`, `fleet`, `host`, something creative
  like Kubernetes did with "kubelet"?
- Should `system hostname` be deprecated/removed?
- Should `system status` move under the new concept (it's really "tell me about
  this node")?
- Should `health status` absorb worker fleet info, or stay purely control-plane?
- How does the worker list output change? Add columns: HOSTNAME, LABELS, STATUS,
  UPTIME, LAST SEEN?

## Notes

- The worker heartbeat registry feature is in progress — workers now register in
  a KV bucket with hostname + labels + timestamp. This enables richer `list`
  output without broadcasting.
- The `changed` field was added to mutation CLI output in #188 but the CLI docs
  (dns/update, exec, shell) were not updated.
- Labels are already configured on workers via `job.worker.labels` in
  `osapi.yaml` and stored in the registry. They just aren't surfaced in the CLI
  output yet.
