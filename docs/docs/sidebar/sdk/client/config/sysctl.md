---
sidebar_position: 1
---

# Sysctl

The `Sysctl` service has methods that manage kernel parameters on target hosts,
through `/etc/sysctl.d/` conf files. Use `client.Sysctl.List()` and
`client.Sysctl.Create()`, for example.

## Methods

| Method                             | Description                           |
| ---------------------------------- | ------------------------------------- |
| `List(ctx, hostname, opts)`        | Every parameter and its current value |
| `Get(ctx, hostname, key)`          | Any parameter's current value         |
| `Create(ctx, hostname, opts)`      | Set a parameter and manage it         |
| `Update(ctx, hostname, key, opts)` | The same, by key                      |
| `Delete(ctx, hostname, key)`       | Stop managing it                      |

## Request Types

| Type               | Fields                           |
| ------------------ | -------------------------------- |
| `SysctlListOpts`   | Prefix, such as `net.ipv4`       |
| `SysctlCreateOpts` | Key (required), Value (required) |
| `SysctlUpdateOpts` | Value (required)                 |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List the parameters under net.ipv4, and which osapi manages
resp, err := c.Sysctl.List(ctx, "web-01", client.SysctlListOpts{Prefix: "net.ipv4"})
for _, entry := range resp.Data.Results {
    fmt.Printf("%s = %s (managed: %t)\n", entry.Key, entry.Value, entry.Managed)
}

// Get a specific parameter
resp, err := c.Sysctl.Get(ctx, "web-01", "net.ipv4.ip_forward")

// Create a new managed parameter
resp, err := c.Sysctl.Create(ctx, "web-01", client.SysctlCreateOpts{
    Key:   "net.ipv4.ip_forward",
    Value: "1",
})

// Update an existing parameter
resp, err := c.Sysctl.Update(ctx, "web-01", "net.ipv4.ip_forward",
    client.SysctlUpdateOpts{
        Value: "0",
    })

// Delete a managed parameter
resp, err := c.Sysctl.Delete(ctx, "web-01", "net.ipv4.ip_forward")
```

## Example

- [`examples/sdk/client/sysctl.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/sysctl.go)

## Permissions

| Operation              | Permission     |
| ---------------------- | -------------- |
| List, Get              | `sysctl:read`  |
| Create, Update, Delete | `sysctl:write` |

Sysctl management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
