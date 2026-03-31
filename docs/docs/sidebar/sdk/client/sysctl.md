---
sidebar_position: 9
---

# Sysctl Service

The `SysctlService` provides methods for managing kernel parameters on target
hosts via `/etc/sysctl.d/` conf files. Access via `client.Sysctl.List()`,
`client.Sysctl.Create()`, etc.

## Methods

| Method                             | Description                 |
| ---------------------------------- | --------------------------- |
| `List(ctx, hostname)`              | List all managed parameters |
| `Get(ctx, hostname, key)`          | Get a parameter by key      |
| `Create(ctx, hostname, opts)`      | Create a new managed entry  |
| `Update(ctx, hostname, key, opts)` | Update an existing entry    |
| `Delete(ctx, hostname, key)`       | Delete a managed entry      |

## Request Types

| Type               | Fields                           |
| ------------------ | -------------------------------- |
| `SysctlCreateOpts` | Key (required), Value (required) |
| `SysctlUpdateOpts` | Value (required)                 |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all managed sysctl entries
resp, err := c.Sysctl.List(ctx, "web-01")
for _, entry := range resp.Data.Results {
    fmt.Printf("%s = %s\n", entry.Key, entry.Value)
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

- [`examples/sdk/client/sysctl.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/sysctl.go)

## Permissions

| Operation              | Permission     |
| ---------------------- | -------------- |
| List, Get              | `sysctl:read`  |
| Create, Update, Delete | `sysctl:write` |

Sysctl management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
