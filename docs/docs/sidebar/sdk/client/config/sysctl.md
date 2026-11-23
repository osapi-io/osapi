---
sidebar_position: 1
---

# Sysctl

The `Sysctl` service has methods that manage kernel parameters on target hosts,
through `/etc/sysctl.d/` conf files. Use `client.Sysctl.List()` and
`client.Sysctl.Create()`, for example.

## Methods

| Method                                                    | Returns                                      | Permission     | Description                                                          |
| --------------------------------------------------------- | -------------------------------------------- | -------------- | -------------------------------------------------------------------- |
| `c.Sysctl.List(ctx, hostname, params SysctlListParams)`   | `Response[Collection[SysctlEntry]]`          | `sysctl:read`  | List the kernel's parameters and which osapi manages                 |
| `c.Sysctl.Create(ctx, hostname, input SysctlCreateInput)` | `Response[Collection[SysctlMutationResult]]` | `sysctl:write` | Manage a kernel parameter and apply it                               |
| `c.Sysctl.Get(ctx, hostname, key)`                        | `Response[Collection[SysctlEntry]]`          | `sysctl:read`  | Get a kernel parameter's current value, and whether osapi manages it |
| `c.Sysctl.Update(ctx, hostname, input SysctlUpdateInput)` | `Response[Collection[SysctlMutationResult]]` | `sysctl:write` | Set a kernel parameter, manage it, and apply it                      |
| `c.Sysctl.Delete(ctx, hostname, key)`                     | `Response[Collection[SysctlMutationResult]]` | `sysctl:write` | Stop the management of a kernel parameter                            |

## Types

The fields of each type are the properties of its schema in the API's spec.

### SysctlListParams

| Field    | Type     | Description                                                  |
| -------- | -------- | ------------------------------------------------------------ |
| `Prefix` | `string` | Only this parameter and the ones under it, such as net.ipv4. |

### SysctlEntry

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Key`      | `string` | Parameter name.                                                                        |
| `Managed`  | `bool`   | Whether osapi manages it: a file of its own sets it at boot.                           |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `Value`    | `string` | Current runtime value.                                                                 |

### SysctlCreateInput

| Field   | Type     | Description                                  |
| ------- | -------- | -------------------------------------------- |
| `Key`   | `string` | Parameter name, such as net.ipv4.ip_forward. |
| `Value` | `string` | Value to set.                                |

### SysctlMutationResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Key`      | `string` | Parameter name.                                                                        |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### SysctlUpdateInput

| Field   | Type     | Description                                                       |
| ------- | -------- | ----------------------------------------------------------------- |
| `Key`   | `string` | The key of the path: parameter name, such as net.ipv4.ip_forward. |
| `Value` | `string` | Value to set.                                                     |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List the parameters under net.ipv4, and which osapi manages
resp, err := c.Sysctl.List(ctx, "web-01", client.SysctlListParams{Prefix: "net.ipv4"})
for _, entry := range resp.Data.Results {
    fmt.Printf("%s = %s (managed: %t)\n", entry.Key, entry.Value, entry.Managed)
}

// Get a specific parameter
resp, err := c.Sysctl.Get(ctx, "web-01", "net.ipv4.ip_forward")

// Create a new managed parameter
resp, err := c.Sysctl.Create(ctx, "web-01", client.SysctlCreateInput{
    Key:   "net.ipv4.ip_forward",
    Value: "1",
})

// Update an existing parameter
resp, err := c.Sysctl.Update(ctx, "web-01", client.SysctlUpdateInput{
    Key: "net.ipv4.ip_forward",
    Value: "0",
})

// Delete a managed parameter
resp, err := c.Sysctl.Delete(ctx, "web-01", "net.ipv4.ip_forward")
```

## Example

- [`examples/sdk/client/sysctl.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/sysctl.go)

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.

Sysctl management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
