---
sidebar_position: 3
---

# Task Functions

Embed custom Go logic in an orchestration plan using `TaskFunc` and
`TaskFuncWithResults`.

## TaskFunc

`TaskFunc` creates a task that runs arbitrary Go code instead of a declarative
operation. The function receives a `context.Context` and the OSAPI
`*client.Client`, and returns a `*Result`.

```go
health := plan.TaskFunc("check-health",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        _, err := c.Health.Liveness(ctx)
        if err != nil {
            return nil, fmt.Errorf("health check: %w", err)
        }

        return &orchestrator.Result{Changed: false}, nil
    },
)
```

Use `TaskFunc` for logic that doesn't map to a built-in operation — health
checks, conditional branching, external API calls, or computed results.

## TaskFuncWithResults

`TaskFuncWithResults` works like `TaskFunc` but also receives a `Results` map
containing outputs from completed upstream tasks. Use it when a task needs data
produced by a prior step.

```go
summary := plan.TaskFuncWithResults("print-summary",
    func(
        _ context.Context,
        _ *client.Client,
        results orchestrator.Results,
    ) (*orchestrator.Result, error) {
        r := results.Get("get-hostname")
        if r == nil {
            return &orchestrator.Result{Changed: false}, nil
        }

        hostname, _ := r.Data["hostname"].(string)
        fmt.Printf("Hostname: %s\n", hostname)

        return &orchestrator.Result{Changed: false}, nil
    },
)
summary.DependsOn(getHostname)
```

`Results.Get(name)` returns the `*Result` for the named task, or `nil` if the
task was not found or has not completed.

## When to Use Each

| Type                  | Use when                                       |
| --------------------- | ---------------------------------------------- |
| `Task` (Op)           | Calling a built-in OSAPI operation             |
| `TaskFunc`            | Running custom logic that doesn't need results |
| `TaskFuncWithResults` | Running logic that reads upstream task data    |

All three types support the same modifiers — `DependsOn`, `When`,
`OnlyIfChanged`, and `OnError`.

## Examples

See
[`examples/sdk/orchestrator/features/task-func.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/task-func.go)
and
[`examples/sdk/orchestrator/features/task-func-results.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/task-func-results.go)
for complete working examples.
