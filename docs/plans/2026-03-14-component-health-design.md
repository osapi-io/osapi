# Component Health and Notifications Design

## Problem

Only agents heartbeat. The API server and NATS server have no presence in the
registry — if they're degraded, the only signal is a failed HTTP call or a NATS
timeout. There's no single view showing all component health, process resource
usage, or condition state. Conditions exist on agents but nothing reacts to
them.

## Goals

1. All three component types (agent, API server, NATS server) heartbeat with
   process metrics and conditions.
2. Health status (`/health/status`) shows a unified component table with TYPE,
   HOSTNAME, STATUS, CONDITIONS, AGE, CPU, MEM.
3. Condition transitions trigger a pluggable notification interface (logging
   stub for now, extensible to Slack/email/webhook later).
4. Remove the `osapi client metrics` CLI command (Prometheus endpoint stays for
   scraping).

## Component Heartbeat

### What gets written

Every component writes a heartbeat to the registry KV bucket on a configurable
interval. The payload includes:

```go
type ComponentRegistration struct {
    // Type is "agent", "api", or "nats".
    Type         string
    Hostname     string
    StartedAt    time.Time
    RegisteredAt time.Time

    // Process metrics — CPU and RSS for the running process.
    Process *ProcessMetrics

    // Conditions — evaluated against thresholds.
    Conditions []Condition

    // Agent-specific fields (nil for api/nats).
    // Labels, OSInfo, LoadAverages, MemoryStats, etc.
    // These remain on AgentRegistration which embeds
    // ComponentRegistration.
}

type ProcessMetrics struct {
    CPUPercent float64
    RSSBytes   int64
    Goroutines int
}
```

Agent heartbeat already collects host-level data (OS, load, memory, disk). That
stays. The new `ProcessMetrics` is added alongside it — CPU/memory for the osapi
process itself, not the host.

### KV key structure

Keep the existing `agent-registry` bucket. Add a type prefix to keys:

```
agent.web-01           → AgentRegistration (embeds ComponentRegistration)
agent.web-02           → AgentRegistration
api.api-server-01      → ComponentRegistration
nats.nats-server-01    → ComponentRegistration
```

The existing key format is `agents.{hostname}`. Changing to `agent.{hostname}`
(no trailing s) is a breaking change. Options:

**Option A**: Keep `agents.` prefix for backward compatibility. Use `api.` and
`nats.` for new component types. `ListAgents` filters by `agents.` prefix.

**Option B**: Migrate to `agent.` prefix. One-time breaking change. Cleaner
going forward.

**Recommendation**: Option A. No migration needed. The prefix inconsistency
(`agents.` vs `agent`) is cosmetic and not worth a breaking change.

### TTL

Same TTL as agent registry (configurable, default 30s). If a component's
heartbeat expires, it disappears from health status — the same liveness
mechanism agents use.

### Collection

Process metrics are collected using Go's `runtime` package and `os.Process()`:

- `runtime.NumGoroutine()` — goroutine count
- `process.MemoryInfo().RSS` — resident set size (via gopsutil or
  /proc/self/status)
- `process.CPUPercent()` — CPU usage since last sample (via gopsutil)

These are cheap calls — safe to run every heartbeat interval.

### Where heartbeat runs

- **Agent**: already has `startHeartbeat()`. Add `ProcessMetrics` to the
  existing `AgentRegistration`.
- **API server**: add `startHeartbeat()` to the API server lifecycle. Writes a
  `ComponentRegistration` with type `"api"`.
- **NATS server**: add `startHeartbeat()` to the NATS server lifecycle. Writes a
  `ComponentRegistration` with type `"nats"`. If the NATS server is external
  (not embedded), this heartbeat doesn't run — and that's fine, the component
  just won't appear in the table.

### Conditions

Agent conditions already exist: `MemoryPressure`, `HighLoad`, `DiskPressure`.
These are host-level.

Add process-level conditions for all components:

- `ProcessMemoryPressure` — process RSS exceeds threshold
- `ProcessHighCPU` — process CPU exceeds threshold

Thresholds are configurable in `osapi.yaml` under each component's config
section. Conditions are evaluated on the component side and written to the
heartbeat — same pattern as agent host conditions.

## Health Status Enrichment

### Component table

`GET /health/status` reads all keys from the registry KV bucket, groups by type
prefix, and returns a component list:

```json
{
  "status": "ok",
  "components": {
    "api": {"status": "ok"},
    "nats": {"status": "ok"},
    "kv": {"status": "ok"}
  },
  "registry": [
    {
      "type": "api",
      "hostname": "api-server-01",
      "status": "Ready",
      "conditions": [],
      "age": "7h 6m",
      "cpu_percent": 2.1,
      "mem_bytes": 134217728
    },
    {
      "type": "nats",
      "hostname": "nats-server-01",
      "status": "Ready",
      "conditions": [],
      "age": "7h 6m",
      "cpu_percent": 0.3,
      "mem_bytes": 67108864
    },
    {
      "type": "agent",
      "hostname": "web-01",
      "status": "Ready",
      "conditions": ["DiskPressure"],
      "age": "7h 6m",
      "cpu_percent": 1.2,
      "mem_bytes": 100663296
    }
  ],
  "jobs": { ... },
  "nats": { ... },
  "streams": [ ... ],
  "kv_buckets": [ ... ]
}
```

The `registry` array replaces the current `agents` field which only shows agent
count/ready. The existing infrastructure sections (jobs, NATS, streams, KV
buckets, object stores, consumers) stay as-is.

### CLI output

`osapi client health status` renders:

```
=== Components ===

TYPE    HOSTNAME                    STATUS  CONDITIONS    AGE    CPU    MEM
api     api-server-01               Ready   -             7h 6m  2.1%   128MB
nats    nats-server-01              Ready   -             7h 6m  0.3%   64MB
agent   Johns-MacBook-Pro-2.local   Ready   DiskPressure  7h 6m  1.2%   96MB
agent   web-02                      Ready   -             3h 2m  0.8%   82MB

=== Jobs ===

  Pending:    2
  Completed:  147
  Failed:     3

=== NATS ===

  URL:      nats://localhost:4222
  Version:  2.10.x
  Streams:  1 (1,234 msgs, 5.2 MB)

=== KV Buckets ===

  job-queue:       42 keys, 1.1 MB
  agent-registry:  2 keys, 4.2 KB
```

Components table at the top — answers "is everything healthy?" at a glance.

## Condition Notifications

### Architecture

The API server watches the registry KV bucket for condition transitions. When a
condition appears or disappears, it dispatches a notification through a
pluggable interface.

```go
// Notifier sends notifications when component conditions change.
type Notifier interface {
    Notify(ctx context.Context, event ConditionEvent) error
}

type ConditionEvent struct {
    ComponentType string    // "agent", "api", "nats"
    Hostname      string
    Condition     string    // "MemoryPressure", "DiskPressure", etc.
    Status        bool      // true = condition active, false = resolved
    Reason        string
    Timestamp     time.Time
}
```

### Watcher

The API server starts a KV watcher on the registry bucket. On each update, it
compares the previous condition set to the current one and emits
`ConditionEvent`s for transitions.

The watcher runs as a background goroutine in the API server lifecycle. It's
designed to be extractable into a separate process later — the only dependency
is NATS KV access and a `Notifier` implementation.

### Implementations

**Phase 1 (this design):**

- `LogNotifier` — logs condition events at INFO level. Default.

**Future phases:**

- `SlackNotifier` — posts to a Slack webhook
- `EmailNotifier` — sends email via SMTP
- `WebhookNotifier` — POSTs to a configurable URL

### Configuration

```yaml
notifications:
  enabled: true
  notifier: log
  # Future:
  # notifier: slack
  # slack:
  #   webhook_url: https://hooks.slack.com/...
  # notifier: webhook
  # webhook:
  #   url: https://example.com/alerts
```

Top-level `notifications` key in `osapi.yaml`. The `notifier` field selects the
implementation. For now only `log` is available.

## Remove `osapi client metrics` CLI

Delete `cmd/client_metrics.go`. The `/metrics` HTTP endpoint stays (Prometheus
scrapes it directly). The CLI command that fetches and prints raw Prometheus
text is not useful for humans.

Also delete:

- `pkg/sdk/client/metrics.go` — SDK MetricsService
- `docs/docs/sidebar/sdk/client/metrics.md` — SDK metrics docs
- CLI docs for the metrics command

The `MetricsService` in the SDK is the only service that bypasses the auth
transport (`http.DefaultClient`). Removing it eliminates that inconsistency.

## What Does NOT Change

- Agent heartbeat data (OS, load, memory, disk, labels, facts) — stays
- Agent conditions (MemoryPressure, HighLoad, DiskPressure) — stays
- `osapi client agent list` output — stays (shows agent-specific detail)
- `osapi client agent get` output — stays
- `/metrics` Prometheus HTTP endpoint — stays
- KV bucket configuration — no new buckets needed
- NATS namespace — no changes

## Order of Implementation

1. Process metrics collection (shared between all components)
2. API server heartbeat
3. NATS server heartbeat (embedded only)
4. Agent heartbeat enrichment (add ProcessMetrics)
5. Health status API changes (registry array, component table)
6. Health status CLI changes (component table output)
7. Condition notification watcher + LogNotifier
8. Configuration (`notifications` section in osapi.yaml)
9. Remove `osapi client metrics` CLI + SDK MetricsService
10. Documentation updates
