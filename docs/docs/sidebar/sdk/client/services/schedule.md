---
sidebar_position: 2
---

# Schedule

The `Schedule` service has methods that manage scheduled entries on target
hosts. On Debian family hosts, these are cron drop-in files. Use
`client.Schedule.List()` and `client.Schedule.Create()`, for example.

## Methods

| Method                                                        | Returns                                | Permission       | Description                                   |
| ------------------------------------------------------------- | -------------------------------------- | ---------------- | --------------------------------------------- |
| `c.Schedule.List(ctx, hostname)`                              | `Response[Collection[ScheduleEntry]]`  | `schedule:read`  | List the cron entries that osapi deployed     |
| `c.Schedule.Create(ctx, hostname, input ScheduleCreateInput)` | `Response[Collection[ScheduleResult]]` | `schedule:write` | Deploy a cron entry                           |
| `c.Schedule.Get(ctx, hostname, name)`                         | `Response[Collection[ScheduleEntry]]`  | `schedule:read`  | Get one cron entry that osapi deployed        |
| `c.Schedule.Update(ctx, hostname, input ScheduleUpdateInput)` | `Response[Collection[ScheduleResult]]` | `schedule:write` | Deploy a cron entry that osapi deployed again |
| `c.Schedule.Delete(ctx, hostname, name)`                      | `Response[Collection[ScheduleResult]]` | `schedule:write` | Remove a cron entry that osapi deployed       |

## Types

The fields of each type are the properties of its schema in the API's spec.

### ScheduleEntry

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Interval` | `string` | Its interval, for a script in /etc/cron.daily and the like.                            |
| `Name`     | `string` | The entry.                                                                             |
| `Object`   | `string` | The object it was deployed from.                                                       |
| `Schedule` | `string` | Its cron expression, for an entry in /etc/cron.d.                                      |
| `Source`   | `string` | Where its file is: cron.d, or the interval.                                            |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `User`     | `string` | The user it runs as.                                                                   |

### ScheduleCreateInput

| Field         | Type                | Description                                                                                                                                        |
| ------------- | ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ContentType` | `string`            | The content type: raw, or template to render the object with the agent's facts and vars; the object's own when empty. It is one of: raw, template. |
| `Interval`    | `string`            | The interval of a script in /etc/cron.daily and the like; give it or schedule. It is one of: hourly, daily, weekly, monthly.                       |
| `Name`        | `string`            | The entry, and its file name, such as backup.                                                                                                      |
| `Object`      | `string`            | The object that holds the entry, from osapi client file upload.                                                                                    |
| `Schedule`    | `string`            | A cron expression, such as */5 * * * *; give it or interval.                                                                                       |
| `User`        | `string`            | The user the entry runs as.                                                                                                                        |
| `Vars`        | `map[string]string` | The vars of a template, as key=value.                                                                                                              |

### ScheduleResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Name`     | `string` | The entry.                                                                             |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### ScheduleUpdateInput

| Field         | Type                | Description                                                                                  |
| ------------- | ------------------- | -------------------------------------------------------------------------------------------- |
| `Name`        | `string`            | The name of the path: the entry, such as backup.                                             |
| `ContentType` | `string`            | The content type, raw or template; the object's own when empty. It is one of: raw, template. |
| `Object`      | `string`            | The object to deploy; the entry's own when empty.                                            |
| `Schedule`    | `string`            | A new cron expression; the entry's own when empty.                                           |
| `User`        | `string`            | The user the entry runs as; the entry's own when empty.                                      |
| `Vars`        | `map[string]string` | The vars of a template, as key=value.                                                        |

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
resp, err := c.Schedule.Create(ctx, "web-01", client.ScheduleCreateInput{
    Name:     "backup-daily",
    Schedule: "0 2 * * *",
    Object:   "backup-script",
    User:     "root",
})

// Create with interval (/etc/cron.daily/)
resp, err := c.Schedule.Create(ctx, "web-01", client.ScheduleCreateInput{
    Name:     "logrotate",
    Interval: "daily",
    Object:   "logrotate-script",
})

// Create with template rendering
resp, err := c.Schedule.Create(ctx, "web-01", client.ScheduleCreateInput{
    Name:        "db-backup",
    Schedule:    "0 4 * * *",
    Object:      "db-backup-template",
    User:        "postgres",
    ContentType: "template",
    Vars:        map[string]string{"db_name": "production"},
})

// Update the schedule and object
resp, err := c.Schedule.Update(ctx, "web-01", client.ScheduleUpdateInput{
    Name: "backup-daily",
    Schedule: "0 3 * * *",
    Object:   "backup-script-v2",
})

// Delete an entry
resp, err := c.Schedule.Delete(ctx, "web-01", "backup-daily")
```

## Example

- [`examples/sdk/client/schedule.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/schedule.go)

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.

Only the `admin` role has `schedule:write`. A cron entry runs a command as its
user, and that user can be root.

Schedule management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On platforms that are not supported (Darwin, generic Linux),
operations return `status: skipped`. Refer to
[Platform Detection](../../platform/detection.md) for details.
