---
sidebar_position: 9
---

# network.ping.do

Ping a host and return latency and packet loss statistics.

## Usage

```go
task := plan.TaskFunc("ping-gateway",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.Ping(ctx, "_any", "192.168.1.1")
        if err != nil {
            return nil, err
        }

        return orchestrator.CollectionResult(
            resp.Data,
            func(r client.PingResult) orchestrator.HostResult {
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

| Param     | Type   | Required | Description                    |
| --------- | ------ | -------- | ------------------------------ |
| `address` | string | Yes      | Hostname or IP address to ping |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Read-only.** Never modifies state. Always returns `Changed: false`.

## Permissions

Requires `network:read` permission.

## Example

See
[`examples/sdk/orchestrator/operations/network-ping.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/network-ping.go)
for a complete working example.
