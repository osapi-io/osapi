---
sidebar_position: 3
---

# Timezone

The `Timezone` service has methods that manage the system timezone on target
hosts through `timedatectl`. Use `client.Timezone.Get()` and
`client.Timezone.Update()`, for example.

## Methods

| Method                                                        | Returns                                | Permission       | Description                  |
| ------------------------------------------------------------- | -------------------------------------- | ---------------- | ---------------------------- |
| `c.Timezone.Get(ctx, hostname)`                               | `Response[Collection[TimezoneEntry]]`  | `timezone:read`  | Get the timezone of the host |
| `c.Timezone.Update(ctx, hostname, input TimezoneUpdateInput)` | `Response[Collection[TimezoneResult]]` | `timezone:write` | Set the timezone of the host |

## Types

The fields of each type are the properties of its schema in the API's spec.

### TimezoneEntry

| Field       | Type     | Description                                                                            |
| ----------- | -------- | -------------------------------------------------------------------------------------- |
| `Error`     | `string` | Error message if the agent failed.                                                     |
| `Hostname`  | `string` | Hostname of the agent that produced this result.                                       |
| `Status`    | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `Timezone`  | `string` | The IANA name of the timezone, such as America/New_York.                               |
| `UTCOffset` | `string` | The offset from UTC now, such as -05:00.                                               |

### TimezoneUpdateInput

| Field      | Type     | Description                                                     |
| ---------- | -------- | --------------------------------------------------------------- |
| `Timezone` | `string` | The IANA name of the timezone, such as UTC or America/New_York. |

### TimezoneResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `Timezone` | `string` | The timezone of the host.                                                              |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get the current timezone
resp, err := c.Timezone.Get(ctx, "web-01")
for _, r := range resp.Data.Results {
    fmt.Printf("timezone=%s offset=%s\n", r.Timezone, r.UTCOffset)
}

// Update the timezone
resp, err := c.Timezone.Update(ctx, "web-01", client.TimezoneUpdateInput{
    Timezone: "America/New_York",
})
for _, r := range resp.Data.Results {
    fmt.Printf("changed=%v\n", r.Changed)
}
```

## Example

- [`examples/sdk/client/timezone.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/timezone.go)

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.

Timezone management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
