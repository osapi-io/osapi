---
sidebar_position: 6
---

# Uptime

System uptime query operations.

## Methods

| Method             | Description |
| ------------------ | ----------- |
| `Get(ctx, target)` | Get uptime  |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get uptime from all hosts
resp, err := c.Uptime.Get(ctx, "_all")
for _, r := range resp.Data.Results {
    fmt.Printf("Uptime (%s): %s\n", r.Hostname, r.Uptime)
}
```

## Example

See
[`examples/sdk/client/uptime.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/uptime.go)
for a complete working example.

## Permissions

Requires `node:read` permission.
