---
sidebar_position: 21
---

# Timezone

The `Timezone` service provides methods for managing the system timezone on
target hosts via `timedatectl`. Access via `client.Timezone.Get()`,
`client.Timezone.Update()`.

## Methods

| Method                        | Description         |
| ----------------------------- | ------------------- |
| `Get(ctx, hostname)`          | Get system timezone |
| `Update(ctx, hostname, opts)` | Set system timezone |

## Request Types

| Type                 | Fields              |
| -------------------- | ------------------- |
| `TimezoneUpdateOpts` | Timezone (required) |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get the current timezone
resp, err := c.Timezone.Get(ctx, "web-01")
for _, r := range resp.Data.Results {
    fmt.Printf("timezone=%s offset=%s\n", r.Timezone, r.UTCOffset)
}

// Update the timezone
resp, err := c.Timezone.Update(ctx, "web-01", client.TimezoneUpdateOpts{
    Timezone: "America/New_York",
})
for _, r := range resp.Data.Results {
    fmt.Printf("changed=%v\n", r.Changed)
}
```

## Result Types

### TimezoneResult

| Field     | Type   | Description                      |
| --------- | ------ | -------------------------------- |
| Hostname  | string | Agent hostname                   |
| Status    | string | ok, failed, or skipped           |
| Timezone  | string | IANA timezone name               |
| UTCOffset | string | UTC offset (e.g., "-05:00")      |
| Error     | string | Error message (empty on success) |

### TimezoneMutationResult

| Field    | Type   | Description                      |
| -------- | ------ | -------------------------------- |
| Hostname | string | Agent hostname                   |
| Status   | string | ok, failed, or skipped           |
| Timezone | string | Timezone that was set            |
| Changed  | bool   | Whether state was modified       |
| Error    | string | Error message (empty on success) |

## Example

- [`examples/sdk/client/timezone.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/timezone.go)

## Permissions

| Operation | Permission       |
| --------- | ---------------- |
| Get       | `timezone:read`  |
| Update    | `timezone:write` |

Timezone management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
