---
sidebar_position: 14
---

# node.uptime.get

Get system uptime information.

## Usage

```go
task := plan.TaskFunc("get-uptime",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.Uptime(ctx, "_any")
        if err != nil {
            return nil, err
        }

        return orchestrator.CollectionResult(
            resp.Data,
            func(r client.UptimeResult) orchestrator.HostResult {
                return orchestrator.HostResult{
                    Hostname: r.Hostname,
                    Changed:  r.Changed,
                    Error:    r.Error,
                }
            },
        ), nil
    },
)
```

## Parameters

None.

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Read-only.** Never modifies state. Always returns `Changed: false`.

## Permissions

Requires `node:read` permission.

## Example

See
[`examples/sdk/orchestrator/operations/node-uptime.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/node-uptime.go)
for a complete working example.
