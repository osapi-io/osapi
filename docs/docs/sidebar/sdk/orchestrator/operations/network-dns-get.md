---
sidebar_position: 7
---

# network.dns.get

Get DNS server configuration for a network interface.

## Usage

```go
task := plan.Task("get-dns", &orchestrator.Op{
    Operation: "network.dns.get",
    Target:    "_any",
    Params: map[string]any{
        "interface": "eth0",
    },
})
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
