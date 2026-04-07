---
sidebar_position: 1
---

# Status

Aggregated node status including OS info, disk, memory, load averages, and
uptime.

## Methods

| Method             | Description                               |
| ------------------ | ----------------------------------------- |
| `Get(ctx, target)` | Full node status (OS, disk, memory, load) |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get full status from all hosts
resp, err := c.Status.Get(ctx, "_all")
for _, r := range resp.Data.Results {
    fmt.Printf("%s: uptime=%s\n", r.Hostname, r.Uptime)
    if r.OSInfo != nil {
        fmt.Printf("  OS: %s %s\n", r.OSInfo.Distribution, r.OSInfo.Version)
    }
}
```

## Example

See
[`examples/sdk/client/status.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/status.go)
for a complete working example.

## Permissions

Requires `node:read` permission.
