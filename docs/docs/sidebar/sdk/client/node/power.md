---
sidebar_position: 2
---

# Power

The `Power` service has methods that manage power state on target hosts through
reboot and shutdown operations. Use `client.Power.Reboot()` and
`client.Power.Shutdown()`.

## Methods

| Method                                                      | Returns                             | Permission      | Description        |
| ----------------------------------------------------------- | ----------------------------------- | --------------- | ------------------ |
| `c.Power.Reboot(ctx, hostname, input PowerRebootInput)`     | `Response[Collection[PowerResult]]` | `power:execute` | Reboot the host    |
| `c.Power.Shutdown(ctx, hostname, input PowerShutdownInput)` | `Response[Collection[PowerResult]]` | `power:execute` | Shut the host down |

## Types

The fields of each type are the properties of its schema in the API's spec.

### PowerRebootInput

| Field     | Type     | Description                                                                  |
| --------- | -------- | ---------------------------------------------------------------------------- |
| `Delay`   | `int`    | Seconds before the host goes down. It rounds to whole minutes, at least one. |
| `Message` | `string` | A message to the logged-in users first.                                      |

### PowerResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Action`   | `string` | Reboot or shutdown.                                                                    |
| `Changed`  | `bool`   | Always true: each run changes the host.                                                |
| `Delay`    | `int`    | Seconds before the host goes down.                                                     |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### PowerShutdownInput

| Field     | Type     | Description                                                                  |
| --------- | -------- | ---------------------------------------------------------------------------- |
| `Delay`   | `int`    | Seconds before the host goes down. It rounds to whole minutes, at least one. |
| `Message` | `string` | A message to the logged-in users first.                                      |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Reboot the host immediately
resp, err := c.Power.Reboot(ctx, "web-01", client.PowerRebootInput{})

// Reboot with a 30-second delay and broadcast message
resp, err := c.Power.Reboot(ctx, "web-01", client.PowerRebootInput{
    Delay:   30,
    Message: "Scheduled maintenance reboot",
})
for _, r := range resp.Data.Results {
    fmt.Printf("action=%s delay=%ds changed=%v\n",
        r.Action, r.Delay, r.Changed)
}

// Shut down the host immediately
resp, err := c.Power.Shutdown(ctx, "web-01", client.PowerShutdownInput{})

// Broadcast shutdown to all hosts with a delay
resp, err := c.Power.Shutdown(ctx, "_all", client.PowerShutdownInput{
    Delay: 60,
})
```

## Example

- [`examples/sdk/client/power.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/power.go)

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.

Power management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
