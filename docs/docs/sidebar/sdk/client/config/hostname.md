---
sidebar_position: 2
---

# Hostname

System hostname query and update operations.

## Methods

| Method                                                        | Returns                                | Permission       | Description                  |
| ------------------------------------------------------------- | -------------------------------------- | ---------------- | ---------------------------- |
| `c.Hostname.Get(ctx, hostname)`                               | `Response[Collection[HostnameEntry]]`  | `hostname:read`  | Get the hostname of the host |
| `c.Hostname.Update(ctx, hostname, input HostnameUpdateInput)` | `Response[Collection[HostnameResult]]` | `hostname:write` | Set the hostname of the host |

## Types

The fields of each type are the properties of its schema in the API's spec.

### HostnameEntry

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Name`     | `string` | The hostname of the host.                                                              |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### HostnameUpdateInput

| Field  | Type     | Description                          |
| ------ | -------- | ------------------------------------ |
| `Name` | `string` | The hostname to set, such as web-01. |

### HostnameResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Name`     | `string` | The hostname of the host.                                                              |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get hostname
resp, err := c.Hostname.Get(ctx, "_any")
for _, r := range resp.Data.Results {
    fmt.Printf("Hostname (%s): %s\n", r.Hostname, r.Name)
}

// Update hostname
resp, err := c.Hostname.Update(ctx, "web-01",
    client.HostnameUpdateInput{Name: "new-hostname"})
for _, r := range resp.Data.Results {
    fmt.Printf("%s: changed=%v\n", r.Name, r.Changed)
}
```

`Update` runs on the Debian family only. Other hosts and containers report
`skipped`.

## Example

Refer to
[`examples/sdk/client/hostname.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/hostname.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
