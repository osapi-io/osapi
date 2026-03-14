# Orchestrator Op Layer Design

## Problem

The SDK orchestrator package (`pkg/sdk/orchestrator/`) has a working DAG engine
(plan, task, runner) but is missing the operation dispatch layer that all
examples and docs reference. Specifically:

1. `orchestrator.Op` struct — not defined
2. `plan.Task(name, &Op{...})` method — not implemented
3. `task.Operation()` accessor — not implemented
4. Operation routing for all 23 documented operations — not built

24 example files reference this API and none compile. The only working
domain-specific code is `docker.go` with 8 hardcoded `DockerXxx()` methods
that bypass the `Op` pattern entirely. These should be removed.

## Design

### Op Struct

```go
// Op describes a declarative operation to execute on a target.
type Op struct {
    // Operation is the dotted operation name (e.g. "docker.pull.execute").
    Operation string

    // Target is the agent routing target (hostname, "_any", "_all",
    // or label selector).
    Target string

    // Params holds operation-specific parameters.
    Params map[string]any
}
```

### plan.Task() Method

```go
// Task creates a declarative operation task, adds it to the plan,
// and returns it. The operation is dispatched to the appropriate
// SDK client method at execution time.
func (p *Plan) Task(name string, op *Op) *Task
```

Internally, `Task()` wraps the op in a `TaskFn` closure that:

1. Looks up the operation in a dispatch registry
2. Calls the corresponding SDK client method with the target and params
3. Extracts `Results[0]` from the `Collection` response
4. Returns a `*Result` with `JobID`, `Changed`, and `Data`

### task.Operation() Accessor

```go
// Operation returns the Op for declarative tasks, or nil for
// TaskFunc tasks.
func (t *Task) Operation() *Op
```

Used by hooks (e.g. `hooks.go` example accesses `task.Operation()` in
`BeforeTask`).

### Operation Registry

A package-level dispatch map keyed by operation string. Each entry is a
function that takes `(ctx, client, target, params)` and returns
`(*Result, error)`.

```go
// opHandler executes an operation against the SDK client.
type opHandler func(
    ctx context.Context,
    c *client.Client,
    target string,
    params map[string]any,
) (*Result, error)
```

The registry covers all 23 documented operations:

| Operation                  | SDK Call                    | Category |
| -------------------------- | --------------------------- | -------- |
| `node.hostname.get`        | `c.Node.Hostname()`         | Node     |
| `node.status.get`          | `c.Node.Status()`           | Node     |
| `node.disk.get`            | `c.Node.Disk()`             | Node     |
| `node.memory.get`          | `c.Node.Memory()`           | Node     |
| `node.uptime.get`          | `c.Node.Uptime()`           | Node     |
| `node.load.get`            | `c.Node.Load()`             | Node     |
| `command.exec.execute`     | `c.Node.Exec()`             | Command  |
| `command.shell.execute`    | `c.Node.Shell()`            | Command  |
| `network.dns.get`          | `c.Node.GetDNS()`           | Network  |
| `network.dns.update`       | `c.Node.UpdateDNS()`        | Network  |
| `network.ping.do`          | `c.Node.Ping()`             | Network  |
| `file.deploy.execute`      | `c.Node.FileDeploy()`       | File     |
| `file.status.get`          | `c.Node.FileStatus()`       | File     |
| `file.upload`              | `c.File.Upload()`           | File     |
| `docker.create.execute`    | `c.Docker.Create()`         | Docker   |
| `docker.list.get`          | `c.Docker.List()`           | Docker   |
| `docker.inspect.get`       | `c.Docker.Inspect()`        | Docker   |
| `docker.start.execute`     | `c.Docker.Start()`          | Docker   |
| `docker.stop.execute`      | `c.Docker.Stop()`           | Docker   |
| `docker.remove.execute`    | `c.Docker.Remove()`         | Docker   |
| `docker.exec.execute`      | `c.Docker.Exec()`           | Docker   |
| `docker.pull.execute`      | `c.Docker.Pull()`           | Docker   |

Note: `file.upload` is unique — it uses `c.File.Upload()` which takes an
`io.Reader`, not a target hostname. The handler will need special treatment
or may be excluded from the `Op` pattern (users can always use `TaskFunc`
for operations that don't fit the standard pattern).

### Result Data Conversion

Each handler converts the SDK result type to `map[string]any` for
`Result.Data`. Rather than hand-coding each conversion, use a shared
`structToMap` helper that marshals via JSON struct tags:

```go
func structToMap(v any) map[string]any {
    b, _ := json.Marshal(v)
    var m map[string]any
    _ = json.Unmarshal(b, &m)
    return m
}
```

This keeps `Data` keys consistent with the SDK type's JSON tags (e.g.
`image_id`, `exit_code`) and requires zero per-type maintenance.

### Collection vs Non-Collection Responses

Most operations return `*Response[Collection[T]]` with a `Results` slice
and `JobID`. The handler extracts `Results[0]` for the `Data` map and
`Changed` bool.

Some operations (`file.upload`, `file.deploy`, `file.status`) return
non-Collection responses. Each handler handles its own response shape.

### File Layout

```
pkg/sdk/orchestrator/
    op.go          — Op struct, plan.Task(), structToMap helper
    registry.go    — opHandler type, dispatch map, all 23 handlers
    docker.go      — DELETED (replaced by registry entries)
```

Split into two files: `op.go` for the public API surface and `registry.go`
for the handler implementations. This keeps the registry internal and
easy to extend.

### Changes to Existing Files

- **`task.go`** — add `op *Op` field, `Operation() *Op` accessor
- **`plan.go`** — no changes (Task method goes in `op.go`)
- **`runner.go`** — no changes (tasks are still `TaskFn` under the hood)
- **`docker.go`** — deleted
- **`docker_public_test.go`** — rewritten to test `plan.Task()` with
  docker operations

### Example Updates

- **`container-targeting.go`** — rewrite to use `plan.Task(&Op{...})`
  instead of `plan.DockerPull()` etc. Add docker operation examples to
  `examples/sdk/orchestrator/operations/`.
- **All 24 broken examples** — will compile once `Op` and `plan.Task()`
  are implemented. No content changes needed.

## What Does NOT Change

- **`pkg/sdk/client/`** — the HTTP client layer is clean. No changes.
- **Orchestrator engine** — plan, runner, result, options, hooks all
  stay as-is. The `Op` layer is additive.
- **`plan.TaskFunc()` and `plan.TaskFuncWithResults()`** — still
  available for custom logic that doesn't fit the `Op` pattern (e.g.
  `file.upload` with `io.Reader`, health checks, custom error handling).

## Testing

- Unit tests for `plan.Task()` covering all 23 operations using
  `httptest.Server` mocks (same pattern as the existing docker tests).
- Tests for unknown operation strings (error case).
- Tests for `task.Operation()` accessor.
- Tests for `structToMap` helper.

## Verification

```bash
go build ./...
go build ./examples/sdk/orchestrator/operations/*.go  # each file individually
go build ./examples/sdk/orchestrator/features/*.go    # each file individually
go test ./pkg/sdk/orchestrator/... -cover
just go::vet
```
