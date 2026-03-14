---
sidebar_position: 8
---

# network.dns.update

Update DNS servers for a network interface.

## Usage

```go
task := plan.TaskFunc("update-dns",
    func(
        ctx context.Context,
        c *client.Client,
    ) (*orchestrator.Result, error) {
        resp, err := c.Node.UpdateDNS(
            ctx,
            "_all",
            "eth0",
            []string{"8.8.8.8", "8.8.4.4"},
            nil,
        )
        if err != nil {
            return nil, err
        }

        return orchestrator.CollectionResult(
            resp.Data,
            func(r client.DNSUpdateResult) orchestrator.HostResult {
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

| Param       | Type     | Required | Description            |
| ----------- | -------- | -------- | ---------------------- |
| `interface` | string   | Yes      | Network interface name |
| `servers`   | []string | Yes      | DNS server addresses   |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Idempotent.** Checks current DNS servers before mutating. Returns
`Changed: true` only if the servers were actually updated. Returns
`Changed: false` if the servers already match the desired state.

## Permissions

Requires `network:write` permission.

## Example

See
[`examples/sdk/orchestrator/operations/network-dns-update.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/orchestrator/operations/network-dns-update.go)
for a complete working example.
