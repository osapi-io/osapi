---
sidebar_position: 3
---

# Disk

Disk usage query operations.

## Methods

| Method             | Description    |
| ------------------ | -------------- |
| `Get(ctx, target)` | Get disk usage |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get disk usage from all hosts
resp, err := c.Disk.Get(ctx, "_all")
for _, r := range resp.Data.Results {
    fmt.Printf("Disk (%s):\n", r.Hostname)
    for _, d := range r.Disks {
        fmt.Printf("  %s  total=%d  used=%d  free=%d\n",
            d.Name, d.Total, d.Used, d.Free)
    }
}
```

## Example

See
[`examples/sdk/client/disk.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/disk.go)
for a complete working example.

## Permissions

Requires `node:read` permission.
