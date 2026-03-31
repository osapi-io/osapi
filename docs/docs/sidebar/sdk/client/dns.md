---
sidebar_position: 6
---

# DNS

DNS configuration query and update operations.

## Methods

| Method                                     | Description        |
| ------------------------------------------ | ------------------ |
| `Get(ctx, target, iface)`                  | Get DNS config     |
| `Update(ctx, target, iface, servers, search)` | Update DNS servers |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get DNS configuration for an interface
resp, err := c.DNS.Get(ctx, "web-01", "eth0")
for _, r := range resp.Data.Results {
    fmt.Printf("Servers: %v\n", r.Servers)
    fmt.Printf("Search:  %v\n", r.SearchDomains)
}

// Update DNS servers
resp, err := c.DNS.Update(
    ctx, "web-01", "eth0",
    []string{"8.8.8.8", "8.8.4.4"},
    nil,
)
```

## Example

See
[`examples/sdk/client/dns.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/dns.go)
for a complete working example.

## Permissions

Get requires `network:read`. Update requires `network:write`.
