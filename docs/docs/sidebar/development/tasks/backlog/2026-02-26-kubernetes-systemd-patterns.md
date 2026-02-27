---
title: Kubernetes and systemd inspired patterns
status: backlog
created: 2026-02-26
updated: 2026-02-26
---

## Objective

Adopt proven patterns from Kubernetes and systemd to make OSAPI's node
management feel more mature and operationally familiar. These are ideas
to explore beyond the initial heartbeat enrichment and `node list`/`node get`
work.

## Ideas

### Node Conditions (Kubernetes-inspired)

Kubernetes nodes report conditions like `MemoryPressure`, `DiskPressure`,
`PIDPressure`, and `NetworkUnavailable`. Since the heartbeat already
collects memory and load data, we could derive conditions from thresholds:

- Memory > 90% used -> `MemoryPressure: true`
- Load 1m > 2x CPU count -> `HighLoad: true`
- Disk > 90% used -> `DiskPressure: true` (would need disk in heartbeat
  or a periodic deep scan)

Conditions would be stored in the KV registration and shown in
`node list` / `node get`. They give operators a quick "is anything
wrong?" signal without digging into raw numbers.

### Capacity and Allocatable (Kubernetes-inspired)

Kubernetes tracks what resources a node has vs. what's available for
scheduling. We could track:

- `max_jobs` (configured) vs. `active_jobs` (current count)
- Job slot utilization per agent visible in `node get`
- Could inform smarter job routing (avoid overloaded agents)

### Taints and Tolerations (Kubernetes-inspired)

Kubernetes nodes can be "tainted" to repel workloads unless they
explicitly tolerate the taint. We already have label-based routing, but
taints would add:

- Mark a node as `draining` or `maintenance` so new jobs avoid it
- `NoSchedule` equivalent: agent stays registered but won't receive
  new jobs
- `NoExecute` equivalent: evict running jobs (graceful drain)
- CLI: `osapi node taint --hostname web-01 --key maintenance --effect NoSchedule`

### Node Lifecycle Events (Kubernetes-inspired)

Kubernetes records lifecycle events per node (Joined, BecameReady,
BecameNotReady, etc.). We could store agent lifecycle events in a
dedicated KV bucket:

- "agent started" with timestamp and version
- "agent stopped" (clean shutdown)
- "heartbeat missed" (detected by TTL expiry watcher)
- "agent restarted" (same hostname re-registers)

Visible via `node get --hostname X` or a dedicated
`node events --hostname X` command.

### Consistent Resource Model (Kubernetes-inspired)

Every Kubernetes object has a uniform envelope: `apiVersion`, `kind`,
`metadata` (name, namespace, labels, annotations, creationTimestamp,
uid), `spec`, `status`. We could formalize OSAPI resources similarly:

- Each resource type (node, job, audit entry) gets a consistent
  structure
- `metadata.labels`, `metadata.annotations`, `metadata.createdAt`
  on every resource
- Annotations (separate from labels) for non-routing metadata
- Enables generic tooling: filtering, sorting, field selectors

### Agent States (systemd-inspired)

Systemd units have explicit states: Active, Inactive, Failed,
Activating, Deactivating. Currently we only have "present in KV =
alive". Adding explicit states would enable:

- `Starting` - agent is initializing, not yet processing jobs
- `Ready` - agent is healthy and processing jobs
- `Draining` - agent is shutting down gracefully, finishing in-flight
  jobs but not accepting new ones
- `Stopped` - clean shutdown (deregistered)

State transitions would be visible in the registry and in lifecycle
events.

### Restart Tracking (systemd-inspired)

Systemd tracks restart counts and restart reasons. We could add:

- `restart_count` - how many times the agent process has started for
  this hostname
- `last_restart_reason` - "clean start", "crash recovery", etc.
- Stability signal for fleet health dashboards

### Additional State to Save

- **First-seen timestamp** (`started_at`) distinct from last heartbeat
  (`registered_at`) for true "AGE" display like `kubectl get nodes`
- **Active job count** - how busy the agent is right now
- **Agent binary version** - for fleet version tracking and rolling
  upgrade visibility
- **OS kernel version** - already available from host provider

## Notes

- These are incremental improvements that build on the heartbeat
  enrichment work. Each can be implemented independently.
- Priority should be driven by operational value: conditions and
  capacity tracking are highest value for fleet operators.
- Taints and lifecycle events add complexity but enable sophisticated
  fleet management workflows.
- The consistent resource model is the most ambitious change and would
  touch the most code, but pays off long-term for tooling and API
  consistency.
