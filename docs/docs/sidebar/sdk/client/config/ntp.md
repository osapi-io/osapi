---
sidebar_position: 2
---

# NTP

The `NTP` service has methods that manage the NTP servers of chrony on target
hosts, through the `ntp` behavior. The servers are in
`/etc/chrony/sources.d/osapi-ntp.sources`. Use `client.NTP.Get()` and
`client.NTP.Create()`, for example.

## Methods

| Method                                              | Returns                           | Permission  | Description                                   |
| --------------------------------------------------- | --------------------------------- | ----------- | --------------------------------------------- |
| `c.NTP.Get(ctx, hostname)`                          | `Response[Collection[NtpStatus]]` | `ntp:read`  | Get the sync status of chrony and its sources |
| `c.NTP.Create(ctx, hostname, input NTPCreateInput)` | `Response[Collection[NtpResult]]` | `ntp:write` | Set the NTP servers that chrony uses          |
| `c.NTP.Update(ctx, hostname, input NTPUpdateInput)` | `Response[Collection[NtpResult]]` | `ntp:write` | Replace the NTP servers that osapi set        |
| `c.NTP.Delete(ctx, hostname)`                       | `Response[Collection[NtpResult]]` | `ntp:write` | Remove the NTP servers that osapi set         |

## Types

The fields of each type are the properties of its schema in the API's spec.

### NtpStatus

| Field           | Type       | Description                                                                            |
| --------------- | ---------- | -------------------------------------------------------------------------------------- |
| `CurrentSource` | `string`   | The source that chrony follows now.                                                    |
| `Error`         | `string`   | Error message if the agent failed.                                                     |
| `Hostname`      | `string`   | Hostname of the agent that produced this result.                                       |
| `Offset`        | `string`   | The offset of the system clock from NTP time, such as +0.000003422s.                   |
| `Servers`       | `[]string` | The sources that chrony has.                                                           |
| `Status`        | `string`   | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `Stratum`       | `int`      | The NTP stratum of the host.                                                           |
| `Synchronized`  | `bool`     | Whether chrony is synchronized to a source.                                            |

### NTPCreateInput

| Field     | Type       | Description                                 |
| --------- | ---------- | ------------------------------------------- |
| `Servers` | `[]string` | The NTP servers, hostnames or IP addresses. |

### NtpResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### NTPUpdateInput

| Field     | Type       | Description                                 |
| --------- | ---------- | ------------------------------------------- |
| `Servers` | `[]string` | The NTP servers, hostnames or IP addresses. |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get NTP status and configured servers
resp, err := c.NTP.Get(ctx, "web-01")
for _, r := range resp.Data.Results {
    fmt.Printf("synchronized=%v stratum=%d source=%s\n",
        r.Synchronized, r.Stratum, r.CurrentSource)
}

// Create NTP configuration
resp, err := c.NTP.Create(ctx, "web-01", client.NTPCreateInput{
    Servers: []string{"0.pool.ntp.org", "1.pool.ntp.org"},
})

// Update NTP servers
resp, err := c.NTP.Update(ctx, "web-01", client.NTPUpdateInput{
    Servers: []string{"ntp.example.com"},
})

// Delete NTP configuration
resp, err := c.NTP.Delete(ctx, "web-01")
```

## Example

- [`examples/sdk/client/ntp.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/ntp.go)

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.

NTP management is supported on the Debian OS family (Ubuntu, Debian, Raspbian).
On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
