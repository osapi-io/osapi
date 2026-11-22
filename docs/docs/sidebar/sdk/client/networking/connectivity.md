---
sidebar_position: 2
---

# Connectivity

Whether a host reaches an address, by ICMP ping: the connectivity behavior.

## Methods

| Method                       | Description                                  |
| ---------------------------- | -------------------------------------------- |
| `Ping(ctx, target, address)` | Ping an address from each host of the target |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Ping an address from all agents
resp, err := c.Connectivity.Ping(ctx, "_all", "8.8.8.8")
for _, r := range resp.Data.Results {
    fmt.Printf("Ping (%s): sent=%d received=%d loss=%.1f%%\n",
        r.Hostname, r.PacketsSent, r.PacketsReceived, r.PacketLoss)
}
```

A host that does not answer is 100% loss, not an error.

## Example

Refer to
[`examples/sdk/client/connectivity.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/connectivity.go)
for a complete working example.

## Permissions

Requires `network:write` permission.
