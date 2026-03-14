---
sidebar_position: 8
---

# Broadcast Targeting

Send operations to multiple agents and access per-host results.

## Targets

| Target      | Behavior                         |
| ----------- | -------------------------------- |
| `_any`      | Any single agent (load balanced) |
| `_all`      | Every registered agent           |
| `hostname`  | Specific host                    |
| `key:value` | Agents matching a label          |

`_all` and label selectors (`key:value`) are broadcast targets — the job runs on
every matching agent and per-host results are available via `HostResults`.

## Usage

```go
getAll := plan.TaskFunc("get-hostname-all",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.Hostname(ctx, "_all")
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

// Access per-host results via TaskFuncWithResults.
printHosts := plan.TaskFuncWithResults("print-hosts",
    func(
        _ context.Context,
        _ *client.Client,
        results orchestrator.Results,
    ) (*orchestrator.Result, error) {
        r := results.Get("get-hostname-all")
        for _, hr := range r.HostResults {
            fmt.Printf("  %s changed=%v\n",
                hr.Hostname, hr.Changed)
        }
        return &orchestrator.Result{Changed: false}, nil
    },
)
printHosts.DependsOn(getAll)
```

## Example

See
[`examples/sdk/orchestrator/features/broadcast.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/features/broadcast.go)
for a complete working example.
