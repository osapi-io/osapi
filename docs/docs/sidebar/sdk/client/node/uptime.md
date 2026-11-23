---
sidebar_position: 6
---

# Uptime

System uptime query operations.

## Methods

| Method                        | Returns                             | Permission    | Description                |
| ----------------------------- | ----------------------------------- | ------------- | -------------------------- |
| `c.Uptime.Get(ctx, hostname)` | `Response[Collection[UptimeEntry]]` | `uptime:read` | Get the uptime of the host |

## Types

The fields of each type are the properties of its schema in the API's spec.

### UptimeEntry

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Seconds`  | `int`    | How long the host has run, in seconds.                                                 |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `Uptime`   | `string` | How long the host has run, in days, hours and minutes.                                 |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get uptime from all hosts
resp, err := c.Uptime.Get(ctx, "_all")
for _, r := range resp.Data.Results {
    fmt.Printf("Uptime (%s): %s\n", r.Hostname, r.Uptime)
}
```

## Example

Refer to
[`examples/sdk/client/uptime.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/uptime.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
