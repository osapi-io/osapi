---
sidebar_position: 2
---

# NTP

The `NTP` service provides methods for managing NTP configuration on target
hosts via chrony drop-in files. Access via `client.NTP.Get()`,
`client.NTP.Create()`, etc.

## Methods

| Method                        | Description                         |
| ----------------------------- | ----------------------------------- |
| `Get(ctx, hostname)`          | Get NTP sync status and server list |
| `Create(ctx, hostname, opts)` | Create NTP configuration            |
| `Update(ctx, hostname, opts)` | Update NTP configuration            |
| `Delete(ctx, hostname)`       | Delete NTP configuration            |

## Request Types

| Type            | Fields             |
| --------------- | ------------------ |
| `NtpCreateOpts` | Servers (required) |
| `NtpUpdateOpts` | Servers (required) |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get NTP status and configured servers
resp, err := c.NTP.Get(ctx, "web-01")
for _, r := range resp.Data.Results {
    fmt.Printf("synchronized=%v stratum=%d source=%s\n",
        r.Synchronized, r.Stratum, r.CurrentSource)
}

// Create NTP configuration
resp, err := c.NTP.Create(ctx, "web-01", client.NtpCreateOpts{
    Servers: []string{"0.pool.ntp.org", "1.pool.ntp.org"},
})

// Update NTP servers
resp, err := c.NTP.Update(ctx, "web-01", client.NtpUpdateOpts{
    Servers: []string{"ntp.example.com"},
})

// Delete NTP configuration
resp, err := c.NTP.Delete(ctx, "web-01")
```

## Example

- [`examples/sdk/client/ntp.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/ntp.go)

## Permissions

| Operation              | Permission  |
| ---------------------- | ----------- |
| Get                    | `ntp:read`  |
| Create, Update, Delete | `ntp:write` |

NTP management is supported on the Debian OS family (Ubuntu, Debian, Raspbian).
On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
