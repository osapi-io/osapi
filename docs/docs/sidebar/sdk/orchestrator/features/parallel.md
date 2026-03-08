---
sidebar_position: 3
---

# Parallel Execution

Tasks at the same DAG level run concurrently. Tasks that share a dependency but
don't depend on each other are automatically parallelized.

## Usage

```go
health := plan.TaskFunc("check-health",
    func(ctx context.Context, c *client.Client) (*orchestrator.Result, error) {
        _, err := c.Health.Liveness(ctx)
        return &orchestrator.Result{Changed: false}, err
    },
)

// Three tasks at the same level — all depend on health,
// so the engine runs them in parallel.
for _, op := range []struct{ name, operation string }{
    {"get-hostname", "node.hostname.get"},
    {"get-disk", "node.disk.get"},
    {"get-memory", "node.memory.get"},
} {
    t := plan.Task(op.name, &orchestrator.Op{
        Operation: op.operation,
        Target:    "_any",
    })
    t.DependsOn(health)
}
```

## Example

See
[`examples/sdk/orchestrator/features/parallel.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/parallel.go)
for a complete working example.
