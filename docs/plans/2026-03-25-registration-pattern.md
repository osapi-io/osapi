# Registration Pattern Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development
> (if subagents available) or superpowers:executing-plans to implement this plan.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate centralized lists that grow with each new component,
replacing them with a provider registry, a simplified 4-method JobClient
interface, and config-driven infrastructure iteration.

**Architecture:** Providers register themselves with a registry at
construction time. The agent dispatches via registry map lookup instead
of a switch statement. The JobClient shrinks from 60+ typed wrapper
methods to 4 generic methods. Config structs expose iteration methods
for KV buckets.

**Tech Stack:** Go 1.25, testify/suite, gomock

---

## Chunk 1: Provider Registry

### Task 1: Create the ProviderRegistry type

**Files:**
- Create: `internal/agent/registry.go`
- Create: `internal/agent/registry_public_test.go`

- [ ] **Step 1: Create registry.go**

```go
package agent

import (
    "encoding/json"
    "fmt"

    "github.com/retr0h/osapi/internal/job"
)

// ProcessorFunc handles job requests for a category.
type ProcessorFunc func(job.Request) (json.RawMessage, error)

// ProviderRegistry holds registered providers and their processors.
type ProviderRegistry struct {
    processors map[string]ProcessorFunc
    providers  []any
}

// NewProviderRegistry creates an empty registry.
func NewProviderRegistry() *ProviderRegistry {
    return &ProviderRegistry{
        processors: make(map[string]ProcessorFunc),
    }
}

// Register adds a category processor and its providers to the registry.
// Providers are collected for FactsAware wiring.
func (r *ProviderRegistry) Register(
    category string,
    processFn ProcessorFunc,
    providers ...any,
) {
    r.processors[category] = processFn
    r.providers = append(r.providers, providers...)
}

// Dispatch routes a job request to the registered processor for its
// category. Returns an error if the category is not registered.
func (r *ProviderRegistry) Dispatch(
    req job.Request,
) (json.RawMessage, error) {
    fn, ok := r.processors[req.Category]
    if !ok {
        return nil, fmt.Errorf(
            "unsupported job category: %s", req.Category)
    }
    return fn(req)
}

// AllProviders returns all registered providers for FactsAware wiring.
func (r *ProviderRegistry) AllProviders() []any {
    return r.providers
}
```

- [ ] **Step 2: Write tests**

Test Register, Dispatch (success + unknown category), AllProviders.

- [ ] **Step 3: Verify**

Run: `go test ./internal/agent/... -count=1`

- [ ] **Step 4: Commit**

```bash
git commit -m "feat: add ProviderRegistry for category-based dispatch"
```

### Task 2: Convert processor methods to standalone functions

Each `processor_*.go` file has methods on `*Agent` that access
`a.xxxProvider`. Convert them to factory functions that return a
`ProcessorFunc` closure capturing the provider.

**Files:**
- Modify: `internal/agent/processor.go`
- Modify: `internal/agent/processor_schedule.go`
- Modify: `internal/agent/processor_docker.go`
- Modify: `internal/agent/processor_file.go`
- Modify: `internal/agent/processor_command.go`

For each processor file, the pattern changes from:

```go
// Before: method on Agent
func (a *Agent) processScheduleOperation(
    jobRequest job.Request,
) (json.RawMessage, error) {
    if a.cronProvider == nil { ... }
    // uses a.cronProvider
}
```

To:

```go
// After: factory returning ProcessorFunc closure
func NewScheduleProcessor(
    cronProvider cron.Provider,
    logger *slog.Logger,
) ProcessorFunc {
    return func(req job.Request) (json.RawMessage, error) {
        if cronProvider == nil { ... }
        // uses cronProvider directly (captured)
    }
}
```

The sub-dispatch functions (processCronList, processCronGet, etc.)
become local functions or closures within the processor.

Do this for all 6 processor files:
- `processor.go` — `processNodeOperation` → `NewNodeProcessor`
  (captures host, disk, mem, load, netinfo providers)
- `processor_schedule.go` — → `NewScheduleProcessor` (captures cron)
- `processor_docker.go` — → `NewDockerProcessor` (captures docker)
- `processor_file.go` — → `NewFileProcessor` (captures file)
- `processor_command.go` — → `NewCommandProcessor` (captures command)
- Network operations in processor.go — → `NewNetworkProcessor`
  (captures dns, ping)

- [ ] **Step 1: Convert processor_schedule.go**

- [ ] **Step 2: Convert processor_docker.go**

- [ ] **Step 3: Convert processor_file.go**

- [ ] **Step 4: Convert processor_command.go**

- [ ] **Step 5: Split processor.go into NewNodeProcessor and
NewNetworkProcessor**

- [ ] **Step 6: Remove the switch in processJobOperation — replace
with registry.Dispatch()**

- [ ] **Step 7: Update tests for all processor files**

The existing tests mock the agent and call processor methods. They
need to call the standalone factory functions instead.

- [ ] **Step 8: Verify all tests pass**

Run: `go test ./internal/agent/... -count=1`

- [ ] **Step 9: Commit**

### Task 3: Simplify Agent struct and constructor

**Files:**
- Modify: `internal/agent/types.go`
- Modify: `internal/agent/agent.go`
- Delete: `internal/agent/factory.go`
- Modify: `cmd/agent_setup.go`

- [ ] **Step 1: Remove provider fields from Agent struct**

Replace 12 provider fields with one `registry *ProviderRegistry`.
Keep `processProvider` (it's telemetry, not a job processor).

- [ ] **Step 2: Simplify agent.New()**

From 17 parameters to ~8:

```go
func New(
    appFs avfs.VFS,
    appConfig config.Config,
    logger *slog.Logger,
    jobClient jobclient.JobClient,
    streamName string,
    registry *ProviderRegistry,
    processProvider process.Provider,
    registryKV jetstream.KeyValue,
    factsKV jetstream.KeyValue,
) *Agent
```

WireProviderFacts uses `registry.AllProviders()`.

- [ ] **Step 3: Delete factory.go**

The factory returned a tuple of providers. Now providers are created
directly in `agent_setup.go` and registered individually.

- [ ] **Step 4: Update agent_setup.go**

```go
registry := agent.NewProviderRegistry()

// Create and register each provider category
hostProv := host.NewDebianProvider()
diskProv := disk.NewDebianProvider(log)
memProv := mem.NewDebianProvider()
loadProv := load.NewDebianProvider()
registry.Register("node",
    agent.NewNodeProcessor(hostProv, diskProv, memProv, loadProv, log),
    hostProv, diskProv, memProv, loadProv)

dnsProv := dns.NewDebianProvider(log, execManager)
pingProv := ping.NewDebianProvider()
registry.Register("network",
    agent.NewNetworkProcessor(dnsProv, pingProv, log),
    dnsProv, pingProv)

commandProv := command.New(log, execManager)
registry.Register("command",
    agent.NewCommandProcessor(commandProv, log),
    commandProv)

// ... file, docker, cron same pattern

a := agent.New(appFs, appConfig, log, jobClient, streamName,
    registry, process.New(), registryKV, factsKV)
```

- [ ] **Step 5: Update all agent test files**

Tests that create agents with mocked providers need to use the
registry pattern instead of passing individual providers.

- [ ] **Step 6: Verify everything**

Run: `go build ./... && go test ./... -count=1 -short`

- [ ] **Step 7: Commit**

---

## Chunk 2: JobClient Interface Simplification

### Task 4: Add generic methods to JobClient

**Files:**
- Modify: `internal/job/client/types.go`
- Modify: `internal/job/client/client.go`

- [ ] **Step 1: Add 4 generic methods to interface**

```go
// Generic job dispatch methods. All typed wrapper methods delegate
// to these. New operations use these directly — no need to add
// methods to the interface.
Query(
    ctx context.Context,
    target string,
    category string,
    operation string,
    data any,
) (string, *job.Response, error)

QueryBroadcast(
    ctx context.Context,
    target string,
    category string,
    operation string,
    data any,
) (string, map[string]*job.Response, map[string]string, error)

Modify(
    ctx context.Context,
    target string,
    category string,
    operation string,
    data any,
) (string, *job.Response, error)

ModifyBroadcast(
    ctx context.Context,
    target string,
    category string,
    operation string,
    data any,
) (string, map[string]*job.Response, map[string]string, error)
```

- [ ] **Step 2: Implement in client.go**

Each method: marshal data → build job.Request → call
publishAndWait/publishAndCollect → process results/errors → return.

- [ ] **Step 3: Write tests for generic methods**

- [ ] **Step 4: Commit**

### Task 5: Migrate API handlers to generic JobClient methods

**Files:**
- Modify: all handler files in `internal/controller/api/node/`
- Modify: all handler files in `internal/controller/api/docker/`
- Modify: all handler files in `internal/controller/api/schedule/`

For each handler, change from:

```go
resp, err := s.JobClient.ModifyDockerCreate(ctx, hostname, data)
```

To:

```go
_, resp, err := s.JobClient.Modify(
    ctx, hostname, "docker", job.OperationDockerCreate, data)
```

And for broadcast:

```go
jobID, results, errs, err := s.JobClient.ModifyBroadcast(
    ctx, target, "docker", job.OperationDockerCreate, data)
```

This is mechanical — same transformation for every handler.

- [ ] **Step 1: Migrate node handlers (7 files)**
- [ ] **Step 2: Migrate docker handlers (9 files)**
- [ ] **Step 3: Migrate schedule handlers (5 files)**
- [ ] **Step 4: Migrate file handlers (3 files)**
- [ ] **Step 5: Migrate command handlers (2 files)**
- [ ] **Step 6: Migrate network handlers (3 files)**
- [ ] **Step 7: Update all handler tests**
- [ ] **Step 8: Verify: `go test ./internal/controller/api/... -count=1`**
- [ ] **Step 9: Commit**

### Task 6: Remove typed wrapper methods

**Files:**
- Delete: `internal/job/client/query.go`
- Delete: `internal/job/client/query_node.go`
- Delete: `internal/job/client/modify.go`
- Delete: `internal/job/client/modify_command.go`
- Delete: `internal/job/client/modify_docker.go`
- Delete: `internal/job/client/schedule_cron.go`
- Delete: `internal/job/client/file.go`
- Modify: `internal/job/client/types.go` — remove old method signatures
- Regenerate: `internal/job/mocks/job_client.gen.go`

- [ ] **Step 1: Remove typed methods from interface**
- [ ] **Step 2: Delete implementation files**
- [ ] **Step 3: Delete typed method test files**
- [ ] **Step 4: Regenerate mocks**
- [ ] **Step 5: Verify: `go build ./... && go test ./... -count=1 -short`**
- [ ] **Step 6: Commit**

---

## Chunk 3: Config-Driven Infrastructure

### Task 7: Add config iteration methods

**Files:**
- Modify: `internal/config/types.go`
- Create: `internal/config/nats.go` (or add to existing)
- Modify: `cmd/controller_setup.go`

- [ ] **Step 1: Add AllKVBucketConfigs method**

Read the NATS config struct first. Add a method that returns all
KV bucket configurations:

```go
func (n NATSConfig) AllKVConfigs() []KVConfig {
    return []KVConfig{
        {Name: "kv", Bucket: n.KV.Bucket, TTL: n.KV.TTL, ...},
        {Name: "response", Bucket: n.KV.ResponseBucket, ...},
        {Name: "audit", Bucket: n.Audit.Bucket, ...},
        {Name: "registry", Bucket: n.Registry.Bucket, ...},
        {Name: "facts", Bucket: n.Facts.Bucket, ...},
        {Name: "state", Bucket: n.State.Bucket, ...},
        {Name: "file_state", Bucket: n.FileState.Bucket, ...},
    }
}
```

- [ ] **Step 2: Update controller_setup.go to iterate**

Replace manual `add()` calls with loop:

```go
for _, cfg := range appConfig.NATS.AllKVConfigs() {
    if cfg.Bucket != "" {
        add(cfg.Bucket)
    }
}
```

- [ ] **Step 3: Write tests**
- [ ] **Step 4: Verify**
- [ ] **Step 5: Commit**

### Task 8: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update provider guide**

Update Step 0 to describe the registry pattern instead of manual
field/parameter wiring. Document:
- How to create a provider and register it
- How to write a NewXxxProcessor factory function
- That adding a provider is ONE change in agent_setup.go
- That JobClient is generic — no new methods needed

- [ ] **Step 2: Commit**

---

## Chunk 4: Verification

### Task 9: Full verification

- [ ] **Step 1:** `go build ./...`
- [ ] **Step 2:** `just go::unit`
- [ ] **Step 3:** `just go::vet`
- [ ] **Step 4:** Verify adding a hypothetical provider requires only
  agent_setup.go + processor file + provider package

---

## Files Summary

| Action | File |
| ------ | ---- |
| Create | `internal/agent/registry.go` |
| Create | `internal/agent/registry_public_test.go` |
| Rewrite | `internal/agent/processor.go` → uses registry.Dispatch |
| Rewrite | `internal/agent/processor_schedule.go` → NewScheduleProcessor |
| Rewrite | `internal/agent/processor_docker.go` → NewDockerProcessor |
| Rewrite | `internal/agent/processor_file.go` → NewFileProcessor |
| Rewrite | `internal/agent/processor_command.go` → NewCommandProcessor |
| Create | `internal/agent/processor_network.go` (split from processor.go) |
| Simplify | `internal/agent/types.go` — 12 fields → 1 registry |
| Simplify | `internal/agent/agent.go` — 17 params → 8 |
| Delete | `internal/agent/factory.go` |
| Rewrite | `cmd/agent_setup.go` — uses registry |
| Add | `internal/job/client/client.go` — 4 generic methods |
| Simplify | `internal/job/client/types.go` — 60→4 interface methods |
| Delete | `internal/job/client/query.go` |
| Delete | `internal/job/client/query_node.go` |
| Delete | `internal/job/client/modify.go` |
| Delete | `internal/job/client/modify_command.go` |
| Delete | `internal/job/client/modify_docker.go` |
| Delete | `internal/job/client/schedule_cron.go` |
| Delete | `internal/job/client/file.go` |
| Modify | All handler files (~29) — use generic JobClient |
| Add | `internal/config/` — AllKVConfigs method |
| Modify | `cmd/controller_setup.go` — iterate config |
| Update | `CLAUDE.md` |
