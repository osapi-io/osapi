---
sidebar_position: 2
---

# Basic Plans

Create a plan, add tasks with dependencies, and run them in order.

## Usage

```go
c := client.New(url, token)
plan := orchestrator.NewPlan(c)

health := plan.TaskFunc("check-health",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        _, err := c.Health.Liveness(ctx)
        return &orchestrator.Result{Changed: false}, err
    },
)

hostname := plan.TaskFunc("get-hostname",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.Hostname(ctx, "_any")
        if err != nil {
            return nil, err
        }

        return orchestrator.CollectionResult(
            resp.Data,
            func(r client.HostnameResult) orchestrator.HostResult {
                return orchestrator.HostResult{
                    Hostname: r.Hostname,
                    Changed:  r.Changed,
                    Error:    r.Error,
                }
            },
        ), nil
    },
)
hostname.DependsOn(health)

report, err := plan.Run(context.Background())
```

`TaskFunc` creates a task with custom Go logic that calls the SDK client
directly. `DependsOn` declares ordering.

## Example

See
[`examples/sdk/orchestrator/features/basic.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/basic.go)
for a complete working example.
