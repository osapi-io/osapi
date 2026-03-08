---
sidebar_position: 4
---

# Guards

Conditional task execution using `When()` predicates and `OnlyIfChanged()`.

## When

`When` takes a predicate that receives completed task results. The task only
runs if the predicate returns `true`.

```go
summary := plan.TaskFunc("print-summary",
    func(_ context.Context, _ *client.Client) (*orchestrator.Result, error) {
        return &orchestrator.Result{Changed: false}, nil
    },
)
summary.DependsOn(getHostname)
summary.When(func(results orchestrator.Results) bool {
    r := results.Get("get-hostname")
    return r != nil && r.Status == orchestrator.StatusChanged
})
```

## WhenWithReason

`WhenWithReason` works like `When` but provides a custom skip reason that is
passed to the `OnSkip` hook when the guard returns `false`.

```go
deploy.WhenWithReason(
    func(results orchestrator.Results) bool {
        r := results.Get("check-config")
        return r != nil && r.Status == orchestrator.StatusChanged
    },
    "config unchanged, skipping deploy",
)
```

Without a reason, skipped tasks report a generic message. Use `WhenWithReason`
when you want descriptive skip output in your hooks.

## OnlyIfChanged

`OnlyIfChanged` skips the task unless at least one dependency reported a change.
See [Only If Changed](only-if-changed.md) for details.

```go
logChange := plan.TaskFunc("log-change",
    func(_ context.Context, _ *client.Client) (*orchestrator.Result, error) {
        return &orchestrator.Result{Changed: true}, nil
    },
)
logChange.DependsOn(deploy)
logChange.OnlyIfChanged()
```

## Example

See
[`examples/sdk/orchestrator/features/guards.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/guards.go)
for a complete working example.
