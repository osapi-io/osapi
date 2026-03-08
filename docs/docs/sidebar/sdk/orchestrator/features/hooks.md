---
sidebar_position: 5
---

# Lifecycle Hooks

Register callbacks at every stage of plan execution. The SDK performs no logging
— hooks are the only output mechanism.

## Hooks

| Hook          | Signature                                       |
| ------------- | ----------------------------------------------- |
| `BeforePlan`  | `func(summary PlanSummary)`                     |
| `AfterPlan`   | `func(report *Report)`                          |
| `BeforeLevel` | `func(level int, tasks []*Task, parallel bool)` |
| `AfterLevel`  | `func(level int, results []TaskResult)`         |
| `BeforeTask`  | `func(task *Task)`                              |
| `AfterTask`   | `func(task *Task, result TaskResult)`           |
| `OnRetry`     | `func(task *Task, attempt int, err error)`      |
| `OnSkip`      | `func(task *Task, reason string)`               |

## Usage

```go
hooks := orchestrator.Hooks{
    BeforePlan: func(summary orchestrator.PlanSummary) {
        fmt.Printf("Plan: %d tasks\n", summary.TotalTasks)
    },
    AfterTask: func(_ *orchestrator.Task, result orchestrator.TaskResult) {
        fmt.Printf("  [%s] %s  changed=%v\n",
            result.Status, result.Name, result.Changed)
    },
}

plan := orchestrator.NewPlan(client, orchestrator.WithHooks(hooks))
```

## Example

See
[`examples/sdk/orchestrator/features/hooks.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/hooks.go)
for a complete working example demonstrating all 8 hooks.
