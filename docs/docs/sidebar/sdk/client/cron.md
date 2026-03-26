---
sidebar_position: 8
---

# Cron Service

The `ScheduleService` provides methods for managing cron drop-in files on target
hosts. Access via `client.Schedule.CronList()`, `client.Schedule.CronCreate()`,
etc.

## Methods

| Method                                  | Description              |
| --------------------------------------- | ------------------------ |
| `CronList(ctx, hostname)`               | List all managed entries |
| `CronGet(ctx, hostname, name)`          | Get entry by name        |
| `CronCreate(ctx, hostname, opts)`       | Create a new entry       |
| `CronUpdate(ctx, hostname, name, opts)` | Update an existing entry |
| `CronDelete(ctx, hostname, name)`       | Delete an entry          |

## Request Types

| Type             | Fields                                                        |
| ---------------- | ------------------------------------------------------------- |
| `CronCreateOpts` | Name, Schedule\*, Interval\*, Command, User (\* one required) |
| `CronUpdateOpts` | Schedule, Command, User (all optional)                        |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all managed cron entries
resp, err := c.Schedule.CronList(ctx, "web-01")
for _, entry := range resp.Data.Results {
    fmt.Printf("%s: %s %s %s\n",
        entry.Name, entry.Schedule, entry.User, entry.Command)
}

// Get a specific entry
resp, err := c.Schedule.CronGet(ctx, "web-01", "backup-daily")

// Create with custom schedule (/etc/cron.d/)
resp, err := c.Schedule.CronCreate(ctx, "web-01", client.CronCreateOpts{
    Name:     "backup-daily",
    Schedule: "0 2 * * *",
    Command:  "/usr/local/bin/backup.sh",
    User:     "root",
})

// Create with interval (/etc/cron.daily/)
resp, err := c.Schedule.CronCreate(ctx, "web-01", client.CronCreateOpts{
    Name:     "logrotate",
    Interval: "daily",
    Command:  "/usr/sbin/logrotate /etc/logrotate.conf",
})

// Update the schedule
resp, err := c.Schedule.CronUpdate(ctx, "web-01", "backup-daily",
    client.CronUpdateOpts{
        Schedule: "0 3 * * *",
    })

// Delete an entry
resp, err := c.Schedule.CronDelete(ctx, "web-01", "backup-daily")
```

## Example

- [`examples/sdk/client/cron.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/cron.go)

## Permissions

| Operation              | Permission   |
| ---------------------- | ------------ |
| List, Get              | `cron:read`  |
| Create, Update, Delete | `cron:write` |

Cron management is supported on the Debian OS family (Ubuntu, Debian, Raspbian).
On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`. See [Platform Detection](../platform/detection.md) for
details.
