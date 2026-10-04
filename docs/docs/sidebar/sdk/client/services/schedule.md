---
sidebar_position: 2
---

# Schedule

The `Schedule` service provides methods for managing scheduled entries on target
hosts. On Debian family hosts these are cron drop-in files. Access via
`client.Schedule.List()`, `client.Schedule.Create()`, etc.

## Methods

| Method                              | Description              |
| ----------------------------------- | ------------------------ |
| `List(ctx, hostname)`               | List all managed entries |
| `Get(ctx, hostname, name)`          | Get entry by name        |
| `Create(ctx, hostname, opts)`       | Create a new entry       |
| `Update(ctx, hostname, name, opts)` | Update an existing entry |
| `Delete(ctx, hostname, name)`       | Delete an entry          |

## Request Types

| Type                 | Fields                                                          |
| -------------------- | --------------------------------------------------------------- |
| `ScheduleCreateOpts` | Name, Object, Schedule\*, Interval\*, User, ContentType, Vars   |
|                      | (\* Schedule and Interval are mutually exclusive; one required) |
| `ScheduleUpdateOpts` | Object, Schedule, User, ContentType, Vars (all optional)        |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all managed entries
resp, err := c.Schedule.List(ctx, "web-01")
for _, entry := range resp.Data.Results {
    fmt.Printf("%s: %s %s %s\n",
        entry.Name, entry.Schedule, entry.User, entry.Object)
}

// Get a specific entry
resp, err := c.Schedule.Get(ctx, "web-01", "backup-daily")

// Create with custom schedule (/etc/cron.d/)
// Object references an uploaded file in the Object Store.
resp, err := c.Schedule.Create(ctx, "web-01", client.ScheduleCreateOpts{
    Name:     "backup-daily",
    Schedule: "0 2 * * *",
    Object:   "backup-script",
    User:     "root",
})

// Create with interval (/etc/cron.daily/)
resp, err := c.Schedule.Create(ctx, "web-01", client.ScheduleCreateOpts{
    Name:     "logrotate",
    Interval: "daily",
    Object:   "logrotate-script",
})

// Create with template rendering
resp, err := c.Schedule.Create(ctx, "web-01", client.ScheduleCreateOpts{
    Name:        "db-backup",
    Schedule:    "0 4 * * *",
    Object:      "db-backup-template",
    User:        "postgres",
    ContentType: "template",
    Vars:        map[string]any{"db_name": "production"},
})

// Update the schedule and object
resp, err := c.Schedule.Update(ctx, "web-01", "backup-daily",
    client.ScheduleUpdateOpts{
        Schedule: "0 3 * * *",
        Object:   "backup-script-v2",
    })

// Delete an entry
resp, err := c.Schedule.Delete(ctx, "web-01", "backup-daily")
```

## Example

- [`examples/sdk/client/schedule.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/schedule.go)

## Permissions

| Operation              | Permission       |
| ---------------------- | ---------------- |
| List, Get              | `schedule:read`  |
| Create, Update, Delete | `schedule:write` |

Schedule management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`. See [Platform Detection](../../platform/detection.md) for
details.
