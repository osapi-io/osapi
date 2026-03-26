# Registration Pattern Design

## Goal

Eliminate centralized lists that grow with each new component. Adding
a new provider, operation, or infrastructure bucket should require
changing 1-2 files, not 5-10.

## Problem

Adding a new provider today touches:

1. `agent/types.go` — add field
2. `agent/agent.go` — add parameter + WireProviderFacts entry
3. `agent/factory.go` — add return value + creation logic
4. `agent/processor.go` — add switch case
5. `cmd/agent_setup.go` — unpack tuple + pass to New()
6. `job/client/types.go` — add 2-4 interface methods
7. `job/client/*.go` — implement methods
8. Regenerate mocks

The JobClient interface has 60+ methods that are thin wrappers around
2 internal functions. KV bucket creation manually lists every bucket
name even though they're all in osapi.yaml.

## Design

### 1. Agent Provider Registry

Replace individual provider fields, parameters, and switch dispatch
with a registry that providers register into at construction time.

**Registry type:**

```go
// internal/agent/registry.go
type ProcessorFunc func(job.Request) (json.RawMessage, error)

type ProviderRegistry struct {
    processors map[string]ProcessorFunc
    providers  []any  // for WireProviderFacts
}

func NewProviderRegistry() *ProviderRegistry {
    return &ProviderRegistry{
        processors: make(map[string]ProcessorFunc),
    }
}

func (r *ProviderRegistry) Register(
    category string,
    provider any,
    processFn ProcessorFunc,
) {
    r.processors[category] = processFn
    r.providers = append(r.providers, provider)
}

func (r *ProviderRegistry) Dispatch(
    req job.Request,
) (json.RawMessage, error) {
    fn, ok := r.processors[req.Category]
    if !ok {
        return nil, fmt.Errorf("unsupported category: %s", req.Category)
    }
    return fn(req)
}

func (r *ProviderRegistry) AllProviders() []any {
    return r.providers
}
```

**Processor functions become standalone closures**, not Agent methods.
They capture their provider dependency:

```go
// internal/agent/processor_schedule.go
func NewScheduleProcessor(
    cronProvider cron.Provider,
    logger *slog.Logger,
) ProcessorFunc {
    return func(req job.Request) (json.RawMessage, error) {
        // dispatch cron sub-operations using cronProvider
    }
}
```

**Agent construction simplifies:**

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
) *Agent {
    a := &Agent{...}
    provider.WireProviderFacts(a.GetFacts, registry.AllProviders()...)
    return a
}
```

**Setup wiring (one file, one place per provider):**

```go
// cmd/agent_setup.go
registry := agent.NewProviderRegistry()

// Node providers
hostProvider := host.NewDebianProvider()
diskProvider := disk.NewDebianProvider(log)
// ...
registry.Register("node", agent.NewNodeProcessor(
    hostProvider, diskProvider, memProvider, loadProvider,
), hostProvider, diskProvider, memProvider, loadProvider)

// Docker
dockerProvider := docker.New()
registry.Register("docker", agent.NewDockerProcessor(
    dockerProvider, log,
), dockerProvider)

// Cron
cronProvider := cron.NewDebianProvider(...)
registry.Register("schedule", agent.NewScheduleProcessor(
    cronProvider, log,
), cronProvider)

agent.New(..., registry, ...)
```

Wait — `Register` takes one provider but some categories have multiple
(node has host, disk, mem, load). The registry needs to accept
multiple providers for FactsAware wiring:

```go
func (r *ProviderRegistry) Register(
    category string,
    processFn ProcessorFunc,
    providers ...any,
) {
    r.processors[category] = processFn
    r.providers = append(r.providers, providers...)
}
```

**What this eliminates:**
- `agent/types.go` — no more provider fields (registry holds them)
- `agent/agent.go` — no more 15-parameter constructor
- `agent/factory.go` — deleted (providers created in setup)
- `agent/processor.go` switch — replaced by registry.Dispatch()
- Each `processor_*.go` — methods on Agent → standalone functions

**What remains:**
- `cmd/agent_setup.go` — still creates providers and registers them
  (this is the ONE place you add a new provider)
- Each `processor_*.go` — still exists as a standalone function

### 2. JobClient Interface Simplification

Replace 60+ typed methods with 4 generic ones:

```go
type JobClient interface {
    Query(
        ctx context.Context,
        target string,
        category string,
        operation string,
        data any,
    ) (*job.Response, error)

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
    ) (*job.Response, error)

    ModifyBroadcast(
        ctx context.Context,
        target string,
        category string,
        operation string,
        data any,
    ) (string, map[string]*job.Response, map[string]string, error)
}
```

**API handlers change from:**
```go
resp, err := s.JobClient.ModifyDockerCreate(ctx, hostname, data)
```

**To:**
```go
resp, err := s.JobClient.Modify(
    ctx, hostname, "docker", job.OperationDockerCreate, data)
```

The operation constants remain typed — `job.OperationDockerCreate` is
still a constant string. Typos are caught by tests, not the compiler.

**What this eliminates:**
- `job/client/types.go` — 60 methods → 4
- `job/client/modify_docker.go`, `query_node.go`, etc. — deleted
  (the generic methods handle all operations)
- Mock regeneration — only 4 methods to mock, never changes
- `job/client/schedule_cron.go`, `modify_command.go`, etc. — deleted

**What remains:**
- `job/client/client.go` — implements the 4 generic methods
- Operation constants — still needed for the string arguments

### 3. Config-Driven Infrastructure

Add methods to the config struct that iterate infrastructure:

```go
// internal/config/nats.go
func (n NATSConfig) AllKVBucketConfigs() []KVBucketConfig {
    return []KVBucketConfig{
        n.KV, n.Audit, n.Registry, n.Facts, n.State, n.FileState,
    }
}

func (n NATSConfig) AllObjectStoreConfigs() []ObjectStoreConfig {
    var configs []ObjectStoreConfig
    if n.Objects.Bucket != "" {
        configs = append(configs, n.Objects)
    }
    return configs
}
```

Then `controller_setup.go` iterates:

```go
for _, cfg := range appConfig.NATS.AllKVBucketConfigs() {
    // create or update bucket
}
```

**What this eliminates:**
- Manual `add(appConfig.NATS.Xxx.Bucket)` calls in setup
- Forgetting to add new buckets to the creation list

**What remains:**
- The config struct still has named fields (needed for typed access)
- `AllKVBucketConfigs()` method needs updating when new buckets are
  added (but it's ONE place, not scattered across setup code)

## Scope

| Change | Files eliminated | Files simplified | New files |
| ------ | --------------- | ---------------- | --------- |
| Provider registry | factory.go deleted | agent.go, types.go, processor.go, setup.go | registry.go |
| JobClient simplification | ~8 typed method files | types.go (60→4 methods), all handler files | None |
| Config iteration | None | controller_setup.go | None |

## What This Does NOT Change

- SDK typed methods — consumers still call `c.Docker.Create()`
- OpenAPI specs — response schemas unchanged
- CLI commands — unchanged
- Operation/permission constants — still explicit declarations
- Provider implementations — unchanged (they don't know about the
  registry)

## Migration Order

1. **Provider registry** first — biggest win, self-contained
2. **JobClient simplification** second — touches many handler files
   but is mechanical (replace method name with generic call)
3. **Config iteration** third — smallest, independent
