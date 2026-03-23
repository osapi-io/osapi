---
sidebar_position: 8
---

# Schedule Service

The `ScheduleService` provides methods for managing cron drop-in files on target
hosts.

## Methods

### CronList

List all osapi-managed cron entries on a target host.

```go
resp, err := c.Schedule.CronList(ctx, "web-01")
for _, entry := range resp.Data.Results {
    fmt.Printf("%s: %s %s %s\n",
        entry.Name, entry.Schedule, entry.User, entry.Command)
}
```

### CronGet

Get a specific cron entry by name.

```go
resp, err := c.Schedule.CronGet(ctx, "web-01", "backup-daily")
fmt.Printf("Schedule: %s\n", resp.Data.Schedule)
```

### CronCreate

Create a new cron drop-in file.

```go
resp, err := c.Schedule.CronCreate(ctx, "web-01", client.CronCreateOpts{
    Name:     "backup-daily",
    Schedule: "0 2 * * *",
    Command:  "/usr/local/bin/backup.sh",
    User:     "root", // optional, defaults to root
})
fmt.Printf("Changed: %v\n", resp.Data.Changed)
```

### CronUpdate

Update an existing cron entry. Only the fields you provide are updated.

```go
resp, err := c.Schedule.CronUpdate(ctx, "web-01", "backup-daily",
    client.CronUpdateOpts{
        Schedule: "0 3 * * *",
    })
fmt.Printf("Changed: %v\n", resp.Data.Changed)
```

### CronDelete

Delete a cron entry.

```go
resp, err := c.Schedule.CronDelete(ctx, "web-01", "backup-daily")
fmt.Printf("Changed: %v\n", resp.Data.Changed)
```

## Types

### CronEntryResult

```go
type CronEntryResult struct {
    Name     string `json:"name"`
    Schedule string `json:"schedule"`
    User     string `json:"user"`
    Command  string `json:"command"`
    Error    string `json:"error,omitempty"`
}
```

### CronCreateOpts

```go
type CronCreateOpts struct {
    Name     string `json:"name"`
    Schedule string `json:"schedule"`
    Command  string `json:"command"`
    User     string `json:"user,omitempty"` // defaults to root
}
```

### CronUpdateOpts

```go
type CronUpdateOpts struct {
    Schedule string `json:"schedule,omitempty"`
    Command  string `json:"command,omitempty"`
    User     string `json:"user,omitempty"`
}
```

### CronMutationResult

```go
type CronMutationResult struct {
    JobID   string `json:"job_id"`
    Name    string `json:"name"`
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}
```

## Supported Platforms

Cron management is supported on the Debian OS family (Ubuntu, Debian, Raspbian).
On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.

## Related

- [Cron Management](../../features/cron-management.md) — feature overview
- [Platform Detection](../platform.md) — OS family detection
