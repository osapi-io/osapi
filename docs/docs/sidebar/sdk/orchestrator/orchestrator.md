---
sidebar_position: 1
---

# Orchestrator

The `orchestrator` package provides DAG-based task orchestration on top of the
OSAPI SDK client. Define tasks with dependencies and the library handles
execution order, parallelism, conditional logic, and reporting.

## Quick Start

```go
import (
    "github.com/retr0h/osapi/pkg/sdk/orchestrator"
    "github.com/retr0h/osapi/pkg/sdk/client"
)

client := client.New("http://localhost:8080", "your-jwt-token")
plan := orchestrator.NewPlan(client)

health := plan.TaskFunc("check-health",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        _, err := c.Health.Liveness(ctx)
        return &orchestrator.Result{Changed: false}, err
    },
)

hostname := plan.Task("get-hostname", &orchestrator.Op{
    Operation: "node.hostname.get",
    Target:    "_any",
})
hostname.DependsOn(health)

report, err := plan.Run(context.Background())
```

## Operations

Operations are the building blocks of orchestration plans. Each operation maps
to an OSAPI job type that agents execute.

| Operation                                                | Description            | Idempotent | Category |
| -------------------------------------------------------- | ---------------------- | ---------- | -------- |
| [`command.exec.execute`](operations/command-exec.md)     | Execute a command      | No         | Command  |
| [`command.shell.execute`](operations/command-shell.md)   | Execute a shell string | No         | Command  |
| [`file.deploy.execute`](operations/file-deploy.md)       | Deploy file to agent   | Yes        | File     |
| [`file.status.get`](operations/file-status.md)           | Check file status      | Read-only  | File     |
| [`file.upload`](operations/file-upload.md)               | Upload to Object Store | Yes        | File     |
| [`network.dns.get`](operations/network-dns-get.md)       | Get DNS configuration  | Read-only  | Network  |
| [`network.dns.update`](operations/network-dns-update.md) | Update DNS servers     | Yes        | Network  |
| [`network.ping.do`](operations/network-ping.md)          | Ping a host            | Read-only  | Network  |
| [`node.hostname.get`](operations/node-hostname.md)       | Get system hostname    | Read-only  | Node     |
| [`node.status.get`](operations/node-status.md)           | Get node status        | Read-only  | Node     |
| [`node.disk.get`](operations/node-disk.md)               | Get disk usage         | Read-only  | Node     |
| [`node.memory.get`](operations/node-memory.md)           | Get memory stats       | Read-only  | Node     |
| [`node.uptime.get`](operations/node-uptime.md)           | Get system uptime      | Read-only  | Node     |
| [`node.load.get`](operations/node-load.md)               | Get load averages      | Read-only  | Node     |
| [`docker.create.execute`](operations/docker-create.md)   | Create a container     | No         | Docker   |
| [`docker.list.get`](operations/docker-list.md)           | List containers        | Read-only  | Docker   |
| [`docker.inspect.get`](operations/docker-inspect.md)     | Inspect a container    | Read-only  | Docker   |
| [`docker.start.execute`](operations/docker-start.md)     | Start a container      | No         | Docker   |
| [`docker.stop.execute`](operations/docker-stop.md)       | Stop a container       | No         | Docker   |
| [`docker.remove.execute`](operations/docker-remove.md)   | Remove a container     | No         | Docker   |
| [`docker.exec.execute`](operations/docker-exec.md)       | Exec in a container    | No         | Docker   |
| [`docker.pull.execute`](operations/docker-pull.md)       | Pull a container image | No         | Docker   |

### Idempotency

- **Read-only** operations never modify state and always return
  `Changed: false`.
- **Idempotent** write operations check current state before mutating and return
  `Changed: true` only if something actually changed.
- **Non-idempotent** operations (command exec/shell) always return
  `Changed: true`. Use guards (`When`, `OnlyIfChanged`) to control when they
  run.

## Hooks

Register callbacks to control logging and progress at every stage:

```go
hooks := orchestrator.Hooks{
    BeforePlan:  func(summary orchestrator.PlanSummary) { ... },
    AfterPlan:   func(report *orchestrator.Report) { ... },
    BeforeLevel: func(level int, tasks []*orchestrator.Task, parallel bool) { ... },
    AfterLevel:  func(level int, results []orchestrator.TaskResult) { ... },
    BeforeTask:  func(task *orchestrator.Task) { ... },
    AfterTask:   func(task *orchestrator.Task, result orchestrator.TaskResult) { ... },
    OnRetry:     func(task *orchestrator.Task, attempt int, err error) { ... },
    OnSkip:      func(task *orchestrator.Task, reason string) { ... },
}

plan := orchestrator.NewPlan(client, orchestrator.WithHooks(hooks))
```

The SDK performs no logging — hooks are the only output mechanism. Consumers
bring their own formatting.

## Error Strategies

| Strategy                          | Behavior                                        |
| --------------------------------- | ----------------------------------------------- |
| `StopAll` (default)               | Fail fast, cancel everything                    |
| `Continue`                        | Skip dependents, keep running independent tasks |
| `Retry(n)`                        | Retry n times immediately before failing        |
| `Retry(n, WithRetryBackoff(...))` | Retry n times with exponential backoff          |

Strategies can be set at plan level or overridden per-task:

```go
plan := orchestrator.NewPlan(client, orchestrator.OnError(orchestrator.Continue))
task.OnError(orchestrator.Retry(3))                                                  // immediate
task.OnError(orchestrator.Retry(3, orchestrator.WithRetryBackoff(1*time.Second, 30*time.Second))) // backoff
```

## Result Types

### Result

The `Result` struct returned by task functions:

| Field         | Type             | Description                               |
| ------------- | ---------------- | ----------------------------------------- |
| `Changed`     | `bool`           | Whether the operation modified state      |
| `Data`        | `map[string]any` | Operation-specific response data          |
| `Status`      | `Status`         | Terminal status (`changed`, `unchanged`)  |
| `HostResults` | `[]HostResult`   | Per-host results for broadcast operations |

### TaskResult

The `TaskResult` struct provided to `AfterTask` hooks and in `Report.Tasks`:

| Field      | Type             | Description                                 |
| ---------- | ---------------- | ------------------------------------------- |
| `Name`     | `string`         | Task name                                   |
| `Status`   | `Status`         | Terminal status                             |
| `Changed`  | `bool`           | Whether the operation reported changes      |
| `Duration` | `time.Duration`  | Execution time                              |
| `Error`    | `error`          | Error if task failed; nil on success        |
| `Data`     | `map[string]any` | Operation response data for post-run access |

### HostResult

Per-host data for broadcast operations (targeting `_all` or label selectors):

| Field      | Type             | Description                        |
| ---------- | ---------------- | ---------------------------------- |
| `Hostname` | `string`         | Agent hostname                     |
| `Changed`  | `bool`           | Whether this host reported changes |
| `Error`    | `string`         | Error message; empty on success    |
| `Data`     | `map[string]any` | Host-specific response data        |

## TaskFuncWithResults

Use `TaskFuncWithResults` when a task needs to read results from prior tasks:

```go
summarize := plan.TaskFuncWithResults(
    "summarize",
    func(
        ctx context.Context,
        client *client.Client,
        results orchestrator.Results,
    ) (*orchestrator.Result, error) {
        r := results.Get("get-hostname")
        hostname := r.Data["hostname"].(string)

        return &orchestrator.Result{
            Changed: true,
            Data:    map[string]any{"summary": hostname},
        }, nil
    },
)
summarize.DependsOn(getHostname)
```

Unlike `TaskFunc`, the function receives the `Results` map containing completed
dependency outputs.

## Features

| Feature                                                | Description                          |
| ------------------------------------------------------ | ------------------------------------ |
| [Basic Plans](features/basic.md)                       | Tasks, dependencies, and execution   |
| [Task Functions](features/task-func.md)                | Custom Go logic with TaskFunc        |
| [Parallel Execution](features/parallel.md)             | Concurrent tasks at the same level   |
| [Guards](features/guards.md)                           | Conditional execution with When      |
| [Only If Changed](features/only-if-changed.md)         | Skip unless dependencies changed     |
| [Lifecycle Hooks](features/hooks.md)                   | Callbacks at every execution stage   |
| [Error Strategies](features/error-strategy.md)         | StopAll, Continue, and Retry         |
| [Failure Recovery](features/only-if-failed.md)         | Recovery tasks on upstream failure   |
| [Retry](features/retry.md)                             | Automatic retry on failure           |
| [Broadcast](features/broadcast.md)                     | Multi-host targeting and HostResults |
| [File Deployment](features/file-deploy-workflow.md)    | Upload, deploy, and verify workflow  |
| [Result Decode](features/result-decode.md)             | Post-run and inter-task data access  |
| [Introspection](features/introspection.md)             | Explain, Levels, and Validate        |
| [Container Targeting](features/container-targeting.md) | Run providers inside containers      |

## Examples

See the
[orchestrator examples](https://github.com/retr0h/osapi/tree/main/examples/sdk/orchestrator/)
for runnable demonstrations of each feature.
