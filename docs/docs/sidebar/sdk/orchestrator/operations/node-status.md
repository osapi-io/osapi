---
sidebar_position: 11
---

# node.status.get

Get comprehensive node status including hostname, OS information, uptime, disk
usage, memory statistics, and load averages.

## Usage

```go
task := plan.TaskFunc("get-status",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.Status(ctx, "web-01")
        if err != nil {
            return nil, err
        }

        return orchestrator.CollectionResult(
            resp.Data,
            func(r client.NodeStatus) orchestrator.HostResult {
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
[`examples/sdk/orchestrator/operations/node-status.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/node-status.go)
for a complete working example.
