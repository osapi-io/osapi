---
sidebar_position: 3
---

# Disk

Disk usage query operations.

## Methods

| Method                      | Returns                           | Permission  | Description                    |
| --------------------------- | --------------------------------- | ----------- | ------------------------------ |
| `c.Disk.Get(ctx, hostname)` | `Response[Collection[DiskEntry]]` | `disk:read` | Get the disk usage of the host |

## Types

The fields of each type are the properties of its schema in the API's spec.

### DiskEntry

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Disks`    | `[]Disk` | Each local filesystem.                                                                 |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

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

Refer to
[`examples/sdk/client/disk.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/disk.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
