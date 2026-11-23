---
sidebar_position: 1
---

# Status

The full node status: OS information, disk, memory, load averages and uptime.

## Methods

| Method                        | Returns                             | Permission    | Description                        |
| ----------------------------- | ----------------------------------- | ------------- | ---------------------------------- |
| `c.Status.Get(ctx, hostname)` | `Response[Collection[StatusEntry]]` | `status:read` | Get the status summary of the host |

## Types

The fields of each type are the properties of its schema in the API's spec.

### StatusEntry

| Field         | Type       | Description                                                                            |
| ------------- | ---------- | -------------------------------------------------------------------------------------- |
| `Disks`       | `[]Disk`   | Each local filesystem.                                                                 |
| `Error`       | `string`   | Error message if the agent failed.                                                     |
| `Errors`      | `[]string` | Each read that failed, as name: error. The other fields still answer.                  |
| `Hostname`    | `string`   | Hostname of the agent that produced this result.                                       |
| `LoadAverage` | `Averages` | LoadAverage is the property load_average.                                              |
| `Memory`      | `Usage`    | Memory is the property memory.                                                         |
| `Name`        | `string`   | The hostname of the host, as the hostname command prints it.                           |
| `OSInfo`      | `Release`  | OSInfo is the property os_info.                                                        |
| `Status`      | `string`   | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `Uptime`      | `Entry`    | Uptime is the property uptime.                                                         |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get full status from all hosts
resp, err := c.Status.Get(ctx, "_all")
for _, r := range resp.Data.Results {
    fmt.Printf("%s: uptime=%s\n", r.Hostname, r.Uptime.Uptime)
    if r.OSInfo.Distribution != "" {
        fmt.Printf("  OS: %s %s\n", r.OSInfo.Distribution, r.OSInfo.Version)
    }
}
```

## Example

Refer to
[`examples/sdk/client/status.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/status.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
