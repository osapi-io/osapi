---
sidebar_position: 7
---

# OS

Operating system information query operations.

## Methods

| Method             | Description               |
| ------------------ | ------------------------- |
| `Get(ctx, target)` | Get operating system info |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get OS info from all hosts
resp, err := c.OS.Get(ctx, "_all")
for _, r := range resp.Data.Results {
    if r.OSInfo != nil {
        fmt.Printf("OS (%s): %s %s\n",
            r.Hostname, r.OSInfo.Distribution, r.OSInfo.Version)
    }
}
```

## Example

See
[`examples/sdk/client/os_info.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/os_info.go)
for a complete working example.

## Permissions

Requires `node:read` permission.
