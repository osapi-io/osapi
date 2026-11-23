---
sidebar_position: 4
---

# Log

`client.Log` reads the systemd journal of target hosts: its entries, and the
sources that wrote them. It calls the `log entry` and `log source` behaviors.

## Methods

| Method                         | What it does                             |
| ------------------------------ | ---------------------------------------- |
| `Entries(ctx, hostname, opts)` | The journal entries that `opts` selects  |
| `Sources(ctx, hostname)`       | Each syslog identifier that has an entry |

Each method returns a `Collection` with one row for each entry, or source, of
each host. A host that skipped or failed has one row with its error.

## Types

| Type              | Fields                                                                           |
| ----------------- | -------------------------------------------------------------------------------- |
| `LogEntryOpts`    | `Unit`, `Lines` (100 when zero), `Since`, `Priority`; each is optional           |
| `LogEntryResult`  | `Hostname`, `Status`, `Timestamp`, `Unit`, `Priority`, `Message`, `PID`, `Error` |
| `LogSourceResult` | `Hostname`, `Status`, `Name`, `Error`                                            |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// The 50 newest entries of web-01.
resp, err := c.Log.Entries(ctx, "web-01", client.LogEntryOpts{Lines: 50})
for _, r := range resp.Data.Results {
    fmt.Printf("[%s] %s %s: %s\n", r.Timestamp, r.Priority, r.Unit, r.Message)
}

// The errors of sshd.service in the last hour, on each host.
resp, err = c.Log.Entries(ctx, "_all", client.LogEntryOpts{
    Unit:     "sshd.service",
    Since:    "1 hour ago",
    Priority: "err",
})

// The sources of the journal.
sources, err := c.Log.Sources(ctx, "web-01")
for _, r := range sources.Data.Results {
    fmt.Println(r.Name)
}
```

## Example

See
[`examples/sdk/client/log.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/log.go)
for a complete program.

## Permissions

`log:read`.
