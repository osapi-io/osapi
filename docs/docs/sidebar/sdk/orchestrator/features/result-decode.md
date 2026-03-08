---
sidebar_position: 10
---

# Result Decode

Access task results after plan execution or pass data between tasks.

## Post-Run Access

After `plan.Run()`, inspect results via `Report.Tasks`:

```go
report, err := plan.Run(context.Background())

for _, r := range report.Tasks {
    fmt.Printf("Task: %s  status=%s  changed=%v\n",
        r.Name, r.Status, r.Changed)
    if len(r.Data) > 0 {
        b, _ := json.MarshalIndent(r.Data, "  ", "  ")
        fmt.Printf("  data=%s\n", b)
    }
}
```

## TaskFuncWithResults

Use `TaskFuncWithResults` to read upstream task data during execution:

```go
summary := plan.TaskFuncWithResults("print-summary",
    func(
        _ context.Context,
        _ *client.Client,
        results orchestrator.Results,
    ) (*orchestrator.Result, error) {
        if r := results.Get("get-hostname"); r != nil {
            if h, ok := r.Data["hostname"].(string); ok {
                fmt.Printf("Hostname: %s\n", h)
            }
        }
        return &orchestrator.Result{Changed: false}, nil
    },
)
summary.DependsOn(getHostname)
```

## Examples

See
[`examples/sdk/orchestrator/features/result-decode.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/result-decode.go)
for a complete working example. For inter-task data passing, see
[Task Functions](task-func.md).
