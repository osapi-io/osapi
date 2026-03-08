---
sidebar_position: 8
---

# Failure Recovery

Trigger recovery or alerting tasks when an upstream task fails.

## Pattern

Combine the `Continue` error strategy with a `When` guard that checks for
`StatusFailed` to build failure-triggered recovery flows.

```go
plan := orchestrator.NewPlan(client,
    orchestrator.OnError(orchestrator.Continue),
)

// A task that might fail.
mightFail := plan.TaskFunc("might-fail",
    func(_ context.Context, _ *client.Client) (*orchestrator.Result, error) {
        return nil, fmt.Errorf("simulated failure")
    },
)
mightFail.OnError(orchestrator.Continue)

// Recovery task — only runs if upstream failed.
alert := plan.TaskFunc("alert",
    func(_ context.Context, _ *client.Client) (*orchestrator.Result, error) {
        fmt.Println("Upstream failed — sending alert!")
        return &orchestrator.Result{Changed: true}, nil
    },
)
alert.DependsOn(mightFail)
alert.When(func(results orchestrator.Results) bool {
    r := results.Get("might-fail")
    return r != nil && r.Status == orchestrator.StatusFailed
})
```

## How It Works

1. The upstream task fails and the `Continue` strategy allows the plan to keep
   running.
2. The downstream task's `When` guard receives the completed `Results` map.
3. The guard checks `r.Status == orchestrator.StatusFailed` — if the upstream
   succeeded, the guard returns `false` and the recovery task is skipped.
4. If the upstream failed, the guard returns `true` and the recovery task
   executes.

Without `Continue`, a failed task with the default `StopAll` strategy would halt
the entire plan before the recovery task gets a chance to run.

## Status Values

| Status            | Meaning                             |
| ----------------- | ----------------------------------- |
| `StatusChanged`   | Task succeeded and reported changes |
| `StatusUnchanged` | Task succeeded with no changes      |
| `StatusSkipped`   | Task was skipped by a guard         |
| `StatusFailed`    | Task failed with an error           |

## Example

See
[`examples/sdk/orchestrator/features/only-if-failed.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/only-if-failed.go)
for a complete working example.
