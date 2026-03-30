---
sidebar_position: 9
---

# Sysctl Service

The `SysctlService` provides methods for managing kernel parameters on target
hosts via `/etc/sysctl.d/` conf files. Access via `client.Sysctl.SysctlList()`,
`client.Sysctl.SysctlCreate()`, etc.

## Methods

| Method                                   | Description                 |
| ---------------------------------------- | --------------------------- |
| `SysctlList(ctx, hostname)`              | List all managed parameters |
| `SysctlGet(ctx, hostname, key)`          | Get a parameter by key      |
| `SysctlCreate(ctx, hostname, opts)`      | Create a new managed entry  |
| `SysctlUpdate(ctx, hostname, key, opts)` | Update an existing entry    |
| `SysctlDelete(ctx, hostname, key)`       | Delete a managed entry      |

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
resp, err := c.Sysctl.SysctlList(ctx, "web-01")
for _, entry := range resp.Data.Results {
    fmt.Printf("%s = %s\n", entry.Key, entry.Value)
}

// Get a specific parameter
resp, err := c.Sysctl.SysctlGet(ctx, "web-01", "net.ipv4.ip_forward")

// Create a new managed parameter
resp, err := c.Sysctl.SysctlCreate(ctx, "web-01", client.SysctlCreateOpts{
    Key:   "net.ipv4.ip_forward",
    Value: "1",
})

// Update an existing parameter
resp, err := c.Sysctl.SysctlUpdate(ctx, "web-01", "net.ipv4.ip_forward",
    client.SysctlUpdateOpts{
        Value: "0",
    })

// Delete a managed parameter
resp, err := c.Sysctl.SysctlDelete(ctx, "web-01", "net.ipv4.ip_forward")
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
