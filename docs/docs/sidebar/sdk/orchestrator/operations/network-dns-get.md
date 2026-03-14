---
sidebar_position: 7
---

# network.dns.get

Get DNS server configuration for a network interface.

## Usage

```go
task := plan.TaskFunc("get-dns",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.GetDNS(ctx, "_any", "eth0")
        if err != nil {
            return nil, err
        }

        return orchestrator.CollectionResult(
            resp.Data,
            func(r client.DNSConfig) orchestrator.HostResult {
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

| Param       | Type   | Required | Description            |
| ----------- | ------ | -------- | ---------------------- |
| `interface` | string | Yes      | Network interface name |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Read-only.** Never modifies state. Always returns `Changed: false`.

## Permissions

Requires `network:read` permission.

## Example

See
[`examples/sdk/orchestrator/operations/network-dns-get.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/network-dns-get.go)
for a complete working example.
