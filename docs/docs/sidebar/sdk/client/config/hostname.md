---
sidebar_position: 2
---

# Hostname

System hostname query and update operations.

## Methods

| Method                      | Description         |
| --------------------------- | ------------------- |
| `Get(ctx, target)`          | Get system hostname |
| `Update(ctx, target, name)` | Set system hostname |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get hostname
resp, err := c.Hostname.Get(ctx, "_any")
for _, r := range resp.Data.Results {
    fmt.Printf("Hostname: %s\n", r.Hostname)
}

// Update hostname
resp, err := c.Hostname.Update(ctx, "web-01", "new-hostname")
for _, r := range resp.Data.Results {
    fmt.Printf("changed=%v\n", r.Changed)
}
```

## Example

See
[`examples/sdk/client/hostname.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/hostname.go)
for a complete working example.

## Permissions

Get requires `node:read`. Update requires `node:write`.
