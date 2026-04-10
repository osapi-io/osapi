---
sidebar_position: 2
---

# Cron

The `Cron` service provides methods for managing cron drop-in files on target
hosts. Access via `client.Cron.List()`, `client.Cron.Create()`, etc.

## Methods

| Method                              | Description              |
| ----------------------------------- | ------------------------ |
| `List(ctx, hostname)`               | List all managed entries |
| `Get(ctx, hostname, name)`          | Get entry by name        |
| `Create(ctx, hostname, opts)`       | Create a new entry       |
| `Update(ctx, hostname, name, opts)` | Update an existing entry |
| `Delete(ctx, hostname, name)`       | Delete an entry          |

## Request Types

| Type             | Fields                                                          |
| ---------------- | --------------------------------------------------------------- |
| `CronCreateOpts` | Name, Object, Schedule\*, Interval\*, User, ContentType, Vars   |
|                  | (\* Schedule and Interval are mutually exclusive; one required) |
| `CronUpdateOpts` | Object, Schedule, User, ContentType, Vars (all optional)        |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all managed cron entries
resp, err := c.Cron.List(ctx, "web-01")
for _, entry := range resp.Data.Results {
    fmt.Printf("%s: %s %s %s\n",
        entry.Name, entry.Schedule, entry.User, entry.Object)
}

// Get a specific entry
resp, err := c.Cron.Get(ctx, "web-01", "backup-daily")

// Create with custom schedule (/etc/cron.d/)
// Object references an uploaded file in the Object Store.
resp, err := c.Cron.Create(ctx, "web-01", client.CronCreateOpts{
    Name:     "backup-daily",
    Schedule: "0 2 * * *",
    Object:   "backup-script",
    User:     "root",
})

// Create with interval (/etc/cron.daily/)
resp, err := c.Cron.Create(ctx, "web-01", client.CronCreateOpts{
    Name:     "logrotate",
    Interval: "daily",
    Object:   "logrotate-script",
})

// Create with template rendering
resp, err := c.Cron.Create(ctx, "web-01", client.CronCreateOpts{
    Name:        "db-backup",
    Schedule:    "0 4 * * *",
    Object:      "db-backup-template",
    User:        "postgres",
    ContentType: "template",
    Vars:        map[string]any{"db_name": "production"},
})

// Update the schedule and object
resp, err := c.Cron.Update(ctx, "web-01", "backup-daily",
    client.CronUpdateOpts{
        Schedule: "0 3 * * *",
        Object:   "backup-script-v2",
    })

// Delete an entry
resp, err := c.Cron.Delete(ctx, "web-01", "backup-daily")
```

## Example

- [`examples/sdk/client/cron.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/cron.go)

## Permissions

| Operation              | Permission   |
| ---------------------- | ------------ |
| List, Get              | `cron:read`  |
| Create, Update, Delete | `cron:write` |

Cron management is supported on the Debian OS family (Ubuntu, Debian, Raspbian).
On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`. See [Platform Detection](../../platform/detection.md) for
details.
