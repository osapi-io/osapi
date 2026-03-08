---
sidebar_position: 9
---

# network.ping.do

Ping a host and return latency and packet loss statistics.

## Usage

```go
task := plan.Task("ping-gateway", &orchestrator.Op{
    Operation: "network.ping.do",
    Target:    "_any",
    Params: map[string]any{
        "address": "192.168.1.1",
    },
})
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
