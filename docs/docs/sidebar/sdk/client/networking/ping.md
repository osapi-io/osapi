---
sidebar_position: 2
---

# Ping

Network ping operations.

## Methods

| Method                     | Description |
| -------------------------- | ----------- |
| `Do(ctx, target, address)` | Ping a host |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Ping a host from all agents
resp, err := c.Ping.Do(ctx, "_all", "8.8.8.8")
for _, r := range resp.Data.Results {
    fmt.Printf("Ping (%s): sent=%d received=%d loss=%.1f%%\n",
        r.Hostname, r.PacketsSent, r.PacketsReceived, r.PacketLoss)
}
```

## Example

See
[`examples/sdk/client/ping.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/ping.go)
for a complete working example.

## Permissions

Requires `network:read` permission.
