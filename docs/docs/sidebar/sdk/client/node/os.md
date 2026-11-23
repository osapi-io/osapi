---
sidebar_position: 7
---

# OS

Operating system information query operations.

## Methods

| Method                    | Returns                         | Permission | Description                          |
| ------------------------- | ------------------------------- | ---------- | ------------------------------------ |
| `c.OS.Get(ctx, hostname)` | `Response[Collection[OsEntry]]` | `os:read`  | Get the operating system of the host |

## Types

The fields of each type are the properties of its schema in the API's spec.

### OsEntry

| Field      | Type      | Description                                                                            |
| ---------- | --------- | -------------------------------------------------------------------------------------- |
| `Error`    | `string`  | Error message if the agent failed.                                                     |
| `Hostname` | `string`  | Hostname of the agent that produced this result.                                       |
| `OSInfo`   | `Release` | OSInfo is the property os_info.                                                        |
| `Status`   | `string`  | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get OS info from all hosts
resp, err := c.OS.Get(ctx, "_all")
for _, r := range resp.Data.Results {
    if r.OSInfo.Distribution != "" {
        fmt.Printf("OS (%s): %s %s\n",
            r.Hostname, r.OSInfo.Distribution, r.OSInfo.Version)
    }
}
```

## Example

Refer to
[`examples/sdk/client/os_info.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/os_info.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
