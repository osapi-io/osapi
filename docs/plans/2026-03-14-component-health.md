# Component Health Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development
> to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make all three components (agent, API server, NATS server) heartbeat
with process metrics and conditions, enrich health status with a unified
component table, add a pluggable condition notification system, and remove the
metrics CLI command.

**Architecture:** Add shared process metrics collection used by all components.
API server and NATS server get heartbeat writers alongside the existing agent
heartbeat. Health status reads all component registrations from KV and renders a
unified table. A KV watcher on the API server dispatches condition transitions
through a pluggable `Notifier` interface.

**Tech Stack:** Go 1.25, gopsutil (process metrics), NATS JetStream KV,
testify/suite

---

## Chunk 1: Process Metrics and Component Registration Types

### Task 1: Add process metrics collector

**Files:**
- Create: `internal/provider/process/process.go`
- Create: `internal/provider/process/types.go`
- Test: `internal/provider/process/process_public_test.go`

- [ ] **Step 1: Define types**

Create `types.go` with:

```go
package process

// Metrics holds process-level resource usage.
type Metrics struct {
    CPUPercent float64 `json:"cpu_percent"`
    RSSBytes   int64   `json:"rss_bytes"`
    Goroutines int     `json:"goroutines"`
}

// Provider collects process metrics.
type Provider interface {
    GetMetrics() (*Metrics, error)
}
```

- [ ] **Step 2: Implement provider**

Create `process.go` using `runtime` and `os` packages. Use
`github.com/shirou/gopsutil/v4/process` (already a dependency via
host/disk/mem/load providers) for CPU% and RSS:

```go
func New() Provider { return &provider{pid: int32(os.Getpid())} }

func (p *provider) GetMetrics() (*Metrics, error) {
    proc, err := gopsutil.NewProcess(p.pid)
    // cpu%, rss, runtime.NumGoroutine()
}
```

- [ ] **Step 3: Write tests**

Test that `GetMetrics` returns non-nil with positive goroutine count
and non-negative CPU/RSS. Use real process — no mocking needed for
a self-inspection provider.

- [ ] **Step 4: Add mockgen**

Create `internal/provider/process/mocks/generate.go`:

```go
//go:generate go tool github.com/golang/mock/mockgen -source=../types.go -destination=types.gen.go -package=mocks
```

Run `go generate ./internal/provider/process/mocks/...`

- [ ] **Step 5: Build and test**

Run: `go build ./...` and `go test ./internal/provider/process/...`

- [ ] **Step 6: Commit**

```
feat: add process metrics provider (CPU, RSS, goroutines)
```

---

### Task 2: Add ComponentRegistration type

**Files:**
- Modify: `internal/job/types.go`

- [ ] **Step 1: Add ComponentRegistration**

Add after `AgentRegistration`:

```go
// ComponentRegistration represents a component's heartbeat entry
// in the KV registry. Used by API server and NATS server.
type ComponentRegistration struct {
    Type         string           `json:"type"`
    Hostname     string           `json:"hostname"`
    StartedAt    time.Time        `json:"started_at"`
    RegisteredAt time.Time        `json:"registered_at"`
    Process      *ProcessMetrics  `json:"process,omitempty"`
    Conditions   []Condition      `json:"conditions,omitempty"`
    Version      string           `json:"version,omitempty"`
}

// ProcessMetrics holds process-level resource usage.
type ProcessMetrics struct {
    CPUPercent float64 `json:"cpu_percent"`
    RSSBytes   int64   `json:"rss_bytes"`
    Goroutines int     `json:"goroutines"`
}
```

- [ ] **Step 2: Add ProcessMetrics to AgentRegistration**

Add field to `AgentRegistration`:

```go
Process *ProcessMetrics `json:"process,omitempty"`
```

This keeps agent-specific fields (OS, load, memory, labels) on
`AgentRegistration` while sharing `ProcessMetrics` with all
component types.

- [ ] **Step 3: Build**

Run: `go build ./...`

- [ ] **Step 4: Commit**

```
feat: add ComponentRegistration and ProcessMetrics types
```

---

## Chunk 2: API Server and NATS Server Heartbeats

### Task 3: Add API server heartbeat

**Files:**
- Create: `internal/api/heartbeat.go`
- Create: `internal/api/heartbeat_test.go`
- Modify: `cmd/api_server_setup.go`

- [ ] **Step 1: Implement API server heartbeat writer**

Create `internal/api/heartbeat.go`:

```go
// StartHeartbeat writes a ComponentRegistration to the registry KV
// on a configurable interval. Call Stop() or cancel the context to
// shut down.
func StartHeartbeat(
    ctx context.Context,
    logger *slog.Logger,
    registryKV jetstream.KeyValue,
    hostname string,
    version string,
    processProvider process.Provider,
    interval time.Duration,
)
```

Key format: `api.{hostname}`. Writes `ComponentRegistration` with
`Type: "api"`. Collects process metrics each tick. Evaluates
process-level conditions (ProcessMemoryPressure, ProcessHighCPU)
using configurable thresholds.

Follow the same pattern as `internal/agent/heartbeat.go`:
- Ticker loop with context cancellation
- Deregister on shutdown (delete KV key)
- Log warnings on errors, don't fail

- [ ] **Step 2: Write tests**

Test in `heartbeat_test.go` (internal test, package `api`):
- Writes registration to mock KV on tick
- Deletes key on context cancel
- Process metrics populated

- [ ] **Step 3: Wire into API server startup**

In `cmd/api_server_setup.go`, after connecting to NATS and getting
the registry KV:
- Create process provider
- Start heartbeat goroutine
- Stop on shutdown

- [ ] **Step 4: Build and test**

Run: `go build ./...` and `go test ./internal/api/...`

- [ ] **Step 5: Commit**

```
feat: add API server heartbeat to component registry
```

---

### Task 4: Add NATS server heartbeat

**Files:**
- Create: `cmd/nats_heartbeat.go`
- Modify: `cmd/nats_setup.go`
- Modify: `cmd/start.go`

- [ ] **Step 1: Implement NATS server heartbeat**

Create `cmd/nats_heartbeat.go` with a heartbeat function similar
to the API server's, but writes with key `nats.{hostname}` and
`Type: "nats"`.

The NATS server heartbeat needs its own NATS client connection
(separate from the server itself) to write to KV. This is created
during `setupJetStream`.

- [ ] **Step 2: Wire into NATS server startup**

In `cmd/nats_setup.go` or `cmd/start.go`, start the NATS heartbeat
after JetStream is set up. Only when running the embedded server
(not when connecting to an external NATS cluster).

- [ ] **Step 3: Build and test**

Run: `go build ./...`

- [ ] **Step 4: Commit**

```
feat: add NATS server heartbeat to component registry
```

---

### Task 5: Add process metrics to agent heartbeat

**Files:**
- Modify: `internal/agent/heartbeat.go`
- Modify: `internal/agent/types.go`
- Modify: `internal/agent/agent.go`
- Modify: `internal/agent/factory.go`

- [ ] **Step 1: Add process provider to agent**

Add `processProvider process.Provider` to `Agent` struct. Initialize
in `factory.go` with `process.New()`. Pass through `New()`.

- [ ] **Step 2: Collect process metrics in heartbeat**

In `writeRegistration`, add:

```go
if pm, err := a.processProvider.GetMetrics(); err == nil {
    reg.Process = &job.ProcessMetrics{
        CPUPercent: pm.CPUPercent,
        RSSBytes:   pm.RSSBytes,
        Goroutines: pm.Goroutines,
    }
}
```

- [ ] **Step 3: Update tests**

Update heartbeat tests to provide a mock process provider.

- [ ] **Step 4: Build and test**

Run: `go build ./...` and `go test ./internal/agent/...`

- [ ] **Step 5: Commit**

```
feat: add process metrics to agent heartbeat
```

---

## Chunk 3: Health Status Enrichment

### Task 6: Update health OpenAPI spec and MetricsProvider

**Files:**
- Modify: `internal/api/health/gen/api.yaml`
- Modify: `internal/api/health/types.go`

- [ ] **Step 1: Add ComponentEntry schema to OpenAPI spec**

Add to the health spec's schemas section:

```yaml
ComponentEntry:
  type: object
  properties:
    type:
      type: string
      description: Component type (agent, api, nats).
    hostname:
      type: string
    status:
      type: string
    conditions:
      type: array
      items:
        type: string
    age:
      type: string
    cpu_percent:
      type: number
    mem_bytes:
      type: integer
      format: int64
```

Add `registry` field to `StatusResponse`:

```yaml
registry:
  type: array
  items:
    $ref: '#/components/schemas/ComponentEntry'
  description: All registered components with health details.
```

- [ ] **Step 2: Regenerate**

Run: `just generate`

- [ ] **Step 3: Add GetComponentRegistry to MetricsProvider**

Add method to `MetricsProvider` interface:

```go
GetComponentRegistry(ctx context.Context) ([]ComponentEntry, error)
```

Add `ComponentEntry` type to `types.go`:

```go
type ComponentEntry struct {
    Type       string
    Hostname   string
    Status     string
    Conditions []string
    Age        string
    CPUPercent float64
    MemBytes   int64
}
```

Update `ClosureMetricsProvider` with `ComponentRegistryFn`.

- [ ] **Step 4: Build**

Run: `go build ./...`

- [ ] **Step 5: Commit**

```
feat: add ComponentEntry to health spec and MetricsProvider
```

---

### Task 7: Implement component registry collection

**Files:**
- Modify: `cmd/api_server_setup.go`
- Modify: `internal/api/health/health_status_get.go`

- [ ] **Step 1: Add ComponentRegistryFn to metrics provider setup**

In `cmd/api_server_setup.go`, add the `ComponentRegistryFn` closure
that reads all keys from the registry KV bucket, parses each as
either `AgentRegistration` or `ComponentRegistration` (based on key
prefix), and returns `[]ComponentEntry`.

Key prefix routing:
- `agents.*` → parse as `AgentRegistration`, type = "agent"
- `api.*` → parse as `ComponentRegistration`, type = "api"
- `nats.*` → parse as `ComponentRegistration`, type = "nats"

For agents, map conditions from the `Conditions` field. For all
types, calculate age from `StartedAt`. Extract CPU/MEM from
`ProcessMetrics`.

- [ ] **Step 2: Add registry to populateMetrics**

In `health_status_get.go`, add a `collect("registry", ...)` call
that runs `GetComponentRegistry` and maps to the response schema.

- [ ] **Step 3: Update tests**

Add test cases for the registry collection — agents + API + NATS
components.

- [ ] **Step 4: Build and test**

Run: `go build ./...` and `go test ./internal/api/health/...`

- [ ] **Step 5: Commit**

```
feat: collect component registry in health status
```

---

### Task 8: Update health status CLI output

**Files:**
- Modify: `cmd/client_health_status.go`

- [ ] **Step 1: Add component table to CLI output**

Render the component registry as a table at the top of the health
status output. Format:

```
=== Components ===

TYPE    HOSTNAME                    STATUS  CONDITIONS    AGE    CPU    MEM
api     api-server-01               Ready   -             7h 6m  2.1%   128MB
nats    nats-server-01              Ready   -             7h 6m  0.3%   64MB
agent   web-01                      Ready   DiskPressure  7h 6m  1.2%   96MB
```

Use `cli.PrintCompactTable` with the component data from the
response. Format CPU as `X.X%`, MEM with `FormatBytes()`.

Keep existing infrastructure sections (Jobs, NATS, Streams, etc.)
below the component table.

- [ ] **Step 2: Build and test manually**

Run: `go build ./...` and test with `go run main.go client health status`

- [ ] **Step 3: Commit**

```
feat: add component table to health status CLI output
```

---

## Chunk 4: Condition Notifications

### Task 9: Add Notifier interface and LogNotifier

**Files:**
- Create: `internal/notify/types.go`
- Create: `internal/notify/log.go`
- Test: `internal/notify/log_public_test.go`

- [ ] **Step 1: Define Notifier interface and ConditionEvent**

```go
package notify

type ConditionEvent struct {
    ComponentType string
    Hostname      string
    Condition     string
    Status        bool      // true = active, false = resolved
    Reason        string
    Timestamp     time.Time
}

type Notifier interface {
    Notify(ctx context.Context, event ConditionEvent) error
}
```

- [ ] **Step 2: Implement LogNotifier**

```go
type LogNotifier struct {
    logger *slog.Logger
}

func NewLogNotifier(logger *slog.Logger) *LogNotifier

func (n *LogNotifier) Notify(ctx context.Context, event ConditionEvent) error
```

Logs at INFO level: `"condition transition"` with structured fields
for component type, hostname, condition, status, reason.

- [ ] **Step 3: Write tests**

Test `LogNotifier.Notify` produces no error. Verify the interface
is satisfied.

- [ ] **Step 4: Commit**

```
feat: add Notifier interface and LogNotifier
```

---

### Task 10: Add condition watcher

**Files:**
- Create: `internal/notify/watcher.go`
- Test: `internal/notify/watcher_test.go`
- Modify: `cmd/api_server_setup.go`

- [ ] **Step 1: Implement KV watcher**

```go
type Watcher struct {
    kv       jetstream.KeyValue
    notifier Notifier
    logger   *slog.Logger
    prev     map[string][]string // hostname → active conditions
}

func NewWatcher(
    kv jetstream.KeyValue,
    notifier Notifier,
    logger *slog.Logger,
) *Watcher

func (w *Watcher) Start(ctx context.Context) error
```

Watches the registry KV bucket. On each update, parses the
registration, compares conditions to `prev`, and emits
`ConditionEvent`s for transitions (new condition → active,
removed condition → resolved).

- [ ] **Step 2: Write tests**

Test condition transition detection:
- New condition appears → Notify called with Status=true
- Condition disappears → Notify called with Status=false
- No change → no notification

- [ ] **Step 3: Wire into API server**

In `cmd/api_server_setup.go`, create the watcher with LogNotifier
and start it as a background goroutine. Stop on shutdown.

- [ ] **Step 4: Add config**

Add to `internal/config/types.go`:

```go
type NotificationsConfig struct {
    Enabled  bool   `mapstructure:"enabled"`
    Notifier string `mapstructure:"notifier"`
}
```

Add `Notifications NotificationsConfig` to `Config` struct.

Update `docs/docs/sidebar/usage/configuration.md` with the new
`notifications` section.

- [ ] **Step 5: Build and test**

Run: `go build ./...` and `go test ./internal/notify/...`

- [ ] **Step 6: Commit**

```
feat: add condition watcher with LogNotifier
```

---

## Chunk 5: Cleanup and Documentation

### Task 11: Remove metrics CLI command

**Files:**
- Delete: `cmd/client_metrics.go`
- Delete: `pkg/sdk/client/metrics.go`
- Delete: `pkg/sdk/client/metrics_public_test.go`
- Delete: `docs/docs/sidebar/sdk/client/metrics.md`
- Modify: `pkg/sdk/client/osapi.go` (remove Metrics field)
- Modify: `docs/docs/sidebar/sdk/client/client.md` (remove Metrics row)

- [ ] **Step 1: Delete files**

Remove the CLI command, SDK service, tests, and docs.

- [ ] **Step 2: Update SDK client**

Remove `Metrics *MetricsService` from `Client` struct and the
initialization in `New()`.

- [ ] **Step 3: Update docs**

Remove Metrics from the SDK client services table.

- [ ] **Step 4: Build and test**

Run: `go build ./...` and `go test ./...`

- [ ] **Step 5: Commit**

```
refactor: remove metrics CLI command and SDK MetricsService

The /metrics Prometheus HTTP endpoint stays for scraping.
The CLI command that prints raw Prometheus text is removed.
```

---

### Task 12: Documentation and verification

**Files:**
- Modify: `docs/docs/sidebar/features/health-checks.md`
- Modify: `docs/docs/sidebar/usage/configuration.md`
- Modify: `docs/docs/sidebar/architecture/system-architecture.md`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update health checks feature doc**

Document the component registry, process metrics, and condition
notifications.

- [ ] **Step 2: Update configuration reference**

Add `notifications` section. Document process condition thresholds.
Add new env vars to the table.

- [ ] **Step 3: Update architecture docs**

Mention component heartbeat in the architecture overview.

- [ ] **Step 4: Full verification**

```bash
go build ./...
go test ./... -count=1
just go::vet
cd docs && bun run build
```

All must pass.

- [ ] **Step 5: Commit**

```
docs: update docs for component health and notifications
```
