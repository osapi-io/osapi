---
sidebar_position: 3
---

# Parallel Execution

Tasks at the same DAG level run concurrently. Tasks that share a dependency but
don't depend on each other are automatically parallelized.

## Usage

```go
health := plan.TaskFunc("check-health",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        _, err := c.Health.Liveness(ctx)
        return &orchestrator.Result{Changed: false}, err
    },
)

// Three tasks at the same level — all depend on health,
// so the engine runs them in parallel.
getHostname := plan.TaskFunc("get-hostname",
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
getHostname.DependsOn(health)

getDisk := plan.TaskFunc("get-disk",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.Disk(ctx, "_any")
        if err != nil {
            return nil, err
        }

        return orchestrator.CollectionResult(
            resp.Data,
            func(r client.DiskResult) orchestrator.HostResult {
                return orchestrator.HostResult{
                    Hostname: r.Hostname,
                    Changed:  r.Changed,
                    Error:    r.Error,
                }
            },
        ), nil
    },
)
getDisk.DependsOn(health)

getMemory := plan.TaskFunc("get-memory",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.Memory(ctx, "_any")
        if err != nil {
            return nil, err
        }

        return orchestrator.CollectionResult(
            resp.Data,
            func(r client.MemoryResult) orchestrator.HostResult {
                return orchestrator.HostResult{
                    Hostname: r.Hostname,
                    Changed:  r.Changed,
                    Error:    r.Error,
                }
            },
        ), nil
    },
)
getMemory.DependsOn(health)
```

## Example

See
[`examples/sdk/orchestrator/features/parallel.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/parallel.go)
for a complete working example.
