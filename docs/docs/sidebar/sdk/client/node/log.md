---
sidebar_position: 4
---

# Log

`client.Log` reads the systemd journal of target hosts: its entries, and the
sources that wrote them. It calls the `log entry` and `log source` behaviors.

## Methods

| Method                                                      | Returns                                 | Permission    | Description                     |
| ----------------------------------------------------------- | --------------------------------------- | ------------- | ------------------------------- |
| `c.LogEntry.List(ctx, hostname, params LogEntryListParams)` | `Response[Collection[LogEntryEntry]]`   | `entry:read`  | List the entries of the journal |
| `c.LogSource.List(ctx, hostname)`                           | `Response[Collection[LogSourceSource]]` | `source:read` | List the sources of the journal |

## Types

The fields of each type are the properties of its schema in the API's spec.

### LogEntryListParams

| Field      | Type     | Description                                                                                      |
| ---------- | -------- | ------------------------------------------------------------------------------------------------ |
| `Unit`     | `string` | Only the entries of this systemd unit, such as nginx.service.                                    |
| `Lines`    | `int`    | How many of the newest entries; 100 when it is not set.                                          |
| `Since`    | `string` | Only the entries since this time, in any form that journalctl --since takes, such as 1 hour ago. |
| `Priority` | `string` | Only the entries at this priority or above.                                                      |

### LogEntryEntry

| Field       | Type     | Description                                                                            |
| ----------- | -------- | -------------------------------------------------------------------------------------- |
| `Error`     | `string` | Error message if the agent failed.                                                     |
| `Hostname`  | `string` | Hostname of the agent that produced this result.                                       |
| `Message`   | `string` | The message.                                                                           |
| `PID`       | `int`    | The process that wrote it.                                                             |
| `Priority`  | `string` | Its priority, by name.                                                                 |
| `Status`    | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `Timestamp` | `string` | When journald recorded it, in RFC 3339.                                                |
| `Unit`      | `string` | The syslog identifier that wrote it.                                                   |

### LogSourceSource

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Name`     | `string` | The syslog identifier, such as sshd.                                                   |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// The 50 newest entries of web-01.
resp, err := c.LogEntry.List(ctx, "web-01", client.LogEntryListParams{Lines: 50})
for _, r := range resp.Data.Results {
    fmt.Printf("[%s] %s %s: %s\n", r.Timestamp, r.Priority, r.Unit, r.Message)
}

// The errors of sshd.service in the last hour, on each host.
resp, err = c.LogEntry.List(ctx, "_all", client.LogEntryListParams{
    Unit:     "sshd.service",
    Since:    "1 hour ago",
    Priority: "err",
})

// The sources of the journal.
sources, err := c.LogSource.List(ctx, "web-01")
for _, r := range sources.Data.Results {
    fmt.Println(r.Name)
}
```

## Example

See
[`examples/sdk/client/log.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/log.go)
for a complete program.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
