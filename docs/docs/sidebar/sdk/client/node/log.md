---
sidebar_position: 4
---

# Log

The `Log` service provides methods for querying the systemd journal on target
hosts. Access via `client.Log`.

## Methods

| Method                                 | Description                               |
| -------------------------------------- | ----------------------------------------- |
| `Query(ctx, hostname, opts)`           | Query journal entries for the host        |
| `QueryUnit(ctx, hostname, unit, opts)` | Query journal entries for a specific unit |
| `Sources(ctx, hostname)`               | List available log sources (syslog IDs)   |

## Request Types

| Type           | Fields                                                        |
| -------------- | ------------------------------------------------------------- |
| `LogQueryOpts` | `Lines` (`*int`), `Since` (`*string`), `Priority` (`*string`) |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Query last 50 journal entries
lines := 50
resp, err := c.Log.Query(ctx, "web-01", client.LogQueryOpts{
    Lines: &lines,
})
for _, r := range resp.Data.Results {
    for _, e := range r.Entries {
        fmt.Printf("[%s] %s %s: %s\n",
            e.Timestamp, e.Priority, e.Unit, e.Message)
    }
}

// Query only error entries from the past hour
since := "1h"
priority := "err"
resp, err := c.Log.Query(ctx, "web-01", client.LogQueryOpts{
    Since:    &since,
    Priority: &priority,
})

// Query entries for a specific systemd unit
resp, err := c.Log.QueryUnit(ctx, "web-01", "sshd.service",
    client.LogQueryOpts{})
for _, r := range resp.Data.Results {
    fmt.Printf("%s: %d entries\n", r.Hostname, len(r.Entries))
    for _, e := range r.Entries {
        fmt.Printf("  [%s] %s\n", e.Priority, e.Message)
    }
}

// List available log sources on the host
srcResp, err := c.Log.Sources(ctx, "web-01")
for _, r := range srcResp.Data.Results {
    for _, src := range r.Sources {
        fmt.Println(src)
    }
}

// Broadcast log query to all hosts
resp, err := c.Log.Query(ctx, "_all", client.LogQueryOpts{})
```

## Result Types

`LogEntryResult` is returned per host in the `Collection.Results` slice:

| Field      | Type         | Description                      |
| ---------- | ------------ | -------------------------------- |
| `Hostname` | `string`     | Target host                      |
| `Status`   | `string`     | `ok`, `skipped`, or `failed`     |
| `Entries`  | `[]LogEntry` | Journal entries (nil if none)    |
| `Error`    | `string`     | Error message if the call failed |

`LogEntry` fields:

| Field       | Type     | Description                         |
| ----------- | -------- | ----------------------------------- |
| `Timestamp` | `string` | ISO 8601 timestamp                  |
| `Unit`      | `string` | Systemd unit name                   |
| `Priority`  | `string` | Log priority (e.g., `info`, `err`)  |
| `Message`   | `string` | Log message text                    |
| `PID`       | `int`    | Process ID that generated the entry |
| `Hostname`  | `string` | Hostname from the journal entry     |

`LogSourceResult` is returned per host for the `Sources` method:

| Field      | Type       | Description                      |
| ---------- | ---------- | -------------------------------- |
| `Hostname` | `string`   | Target host                      |
| `Status`   | `string`   | `ok`, `skipped`, or `failed`     |
| `Sources`  | `[]string` | Syslog identifiers (sorted)      |
| `Error`    | `string`   | Error message if the call failed |

## Example

- [`examples/sdk/client/log.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/log.go)

## Permissions

| Operation                 | Permission |
| ------------------------- | ---------- |
| Query, QueryUnit, Sources | `log:read` |

Log management is supported on the Debian OS family (Ubuntu, Debian, Raspbian).
On unsupported platforms (Darwin, generic Linux) and inside containers,
operations return `status: skipped`.
