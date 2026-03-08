---
sidebar_position: 2
---

# Basic Plans

Create a plan, add tasks with dependencies, and run them in order.

## Usage

```go
client := client.New(url, token)
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

`Task` creates an Op-based task (sends a job to an agent). `TaskFunc` embeds
custom Go logic. `DependsOn` declares ordering.

## Example

See
[`examples/sdk/orchestrator/features/basic.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/basic.go)
for a complete working example.
