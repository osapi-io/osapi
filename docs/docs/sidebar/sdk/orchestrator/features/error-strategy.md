---
sidebar_position: 6
---

# Error Strategies

Control what happens when a task fails.

## Strategies

| Strategy                         | Behavior                                        |
| -------------------------------- | ----------------------------------------------- |
| `StopAll` (default)              | Fail fast, cancel everything                    |
| `Continue`                       | Skip dependents, keep running independent tasks |
| `Retry(n)`                       | Retry n times immediately before failing        |
| `Retry(n, WithRetryBackoff(...))` | Retry n times with exponential backoff          |

## Usage

Set at plan level or override per-task:

```go
// Plan-level: don't halt on failure.
plan := orchestrator.NewPlan(client,
    orchestrator.OnError(orchestrator.Continue),
)

// Task-level override: immediate retry.
task.OnError(orchestrator.Retry(3))

// Task-level override: retry with exponential backoff.
task.OnError(orchestrator.Retry(3,
    orchestrator.WithRetryBackoff(1*time.Second, 30*time.Second),
))
```

With `Continue`, independent tasks keep running when one fails. With `StopAll`,
the entire plan halts on the first failure.

## Failure Recovery

Use a `When` guard to trigger recovery tasks when an upstream fails:

```go
alert := plan.TaskFunc("alert",
    func(_ context.Context, _ *client.Client) (*orchestrator.Result, error) {
        return &orchestrator.Result{Changed: true}, nil
    },
)
alert.DependsOn(mightFail)
alert.When(func(results orchestrator.Results) bool {
    r := results.Get("might-fail")
    return r != nil && r.Status == orchestrator.StatusFailed
})
```

## Examples

See
[`examples/sdk/orchestrator/features/error-strategy.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/error-strategy.go)
for a complete working example. For failure-triggered recovery, see
[Failure Recovery](only-if-failed.md).
