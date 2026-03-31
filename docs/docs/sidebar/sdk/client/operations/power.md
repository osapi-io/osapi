---
sidebar_position: 2
---

# Power

The `Power` service provides methods for managing power state on target hosts
via reboot and shutdown operations. Access via `client.Power.Reboot()` and
`client.Power.Shutdown()`.

## Methods

| Method                          | Description                       |
| ------------------------------- | --------------------------------- |
| `Reboot(ctx, hostname, opts)`   | Schedule a reboot on the target   |
| `Shutdown(ctx, hostname, opts)` | Schedule a shutdown on the target |

## Request Types

| Type        | Fields                                     |
| ----------- | ------------------------------------------ |
| `PowerOpts` | `Delay` (int, seconds), `Message` (string) |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Reboot the host immediately
resp, err := c.Power.Reboot(ctx, "web-01", client.PowerOpts{})

// Reboot with a 30-second delay and broadcast message
resp, err := c.Power.Reboot(ctx, "web-01", client.PowerOpts{
    Delay:   30,
    Message: "Scheduled maintenance reboot",
})
for _, r := range resp.Data.Results {
    fmt.Printf("action=%s delay=%ds changed=%v\n",
        r.Action, r.Delay, r.Changed)
}

// Shut down the host immediately
resp, err := c.Power.Shutdown(ctx, "web-01", client.PowerOpts{})

// Broadcast shutdown to all hosts with a delay
resp, err := c.Power.Shutdown(ctx, "_all", client.PowerOpts{
    Delay: 60,
})
```

## Example

- [`examples/sdk/client/power.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/power.go)

## Permissions

| Operation        | Permission      |
| ---------------- | --------------- |
| Reboot, Shutdown | `power:execute` |

Power management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
