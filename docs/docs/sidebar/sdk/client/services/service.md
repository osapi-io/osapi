---
sidebar_position: 5
---

# Service

Read the systemd services of target hosts. Deploy, remove, start, stop, restart,
enable and disable the services that osapi deploys. The unit file of each is an
object that `File.Upload` stores one time. Refer to
[Service management](../../../features/service-management.md).

## Methods

| Method                              | Description                                            |
| ----------------------------------- | ------------------------------------------------------ |
| `List(ctx, hostname)`               | List the services, one row per service per host        |
| `Get(ctx, hostname, name)`          | Get one service                                        |
| `Create(ctx, hostname, opts)`       | Deploy an object as the unit file `osapi-<name>`       |
| `Update(ctx, hostname, name, opts)` | Deploy a different object over the unit file           |
| `Delete(ctx, hostname, name)`       | Stop, disable and remove a service that osapi deployed |
| `Start(ctx, hostname, name)`        | Start a service that osapi deployed                    |
| `Stop(ctx, hostname, name)`         | Stop a service that osapi deployed                     |
| `Restart(ctx, hostname, name)`      | Restart a service that osapi deployed                  |
| `Enable(ctx, hostname, name)`       | Make it start at boot                                  |
| `Disable(ctx, hostname, name)`      | Make it not start at boot                              |

## Request types

| Type                | Fields                             |
| ------------------- | ---------------------------------- |
| `ServiceCreateOpts` | Name (required), Object (required) |
| `ServiceUpdateOpts` | Object (required)                  |

## Result types

### ServiceInfo (List, Get)

| Field         | Type     | Description                                  |
| ------------- | -------- | -------------------------------------------- |
| `Hostname`    | `string` | Agent hostname                               |
| `Status`      | `string` | Result status (`ok`, `skipped`, `failed`)    |
| `Name`        | `string` | The service, without `.service`              |
| `State`       | `string` | What systemd says it does, such as `running` |
| `Enabled`     | `bool`   | Whether it starts at boot                    |
| `Description` | `string` | The `Description=` of its unit               |
| `Error`       | `string` | Error message (if any)                       |

### ServiceMutationResult (writes and actions)

| Field      | Type     | Description                               |
| ---------- | -------- | ----------------------------------------- |
| `Hostname` | `string` | Agent hostname                            |
| `Status`   | `string` | Result status (`ok`, `skipped`, `failed`) |
| `Name`     | `string` | The service                               |
| `Changed`  | `bool`   | Whether the host changed                  |
| `Error`    | `string` | Error message (if any)                    |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List the services: one row per service per host
resp, err := c.Service.List(ctx, "web-01")
for _, r := range resp.Data.Results {
    fmt.Printf("%s %s state=%s enabled=%v\n",
        r.Hostname, r.Name, r.State, r.Enabled)
}

// Deploy the uploaded object app.service as osapi-app.service, and start it
_, err = c.Service.Create(ctx, "web-01",
    client.ServiceCreateOpts{Name: "app", Object: "app.service"})
_, err = c.Service.Start(ctx, "web-01", "app")

// Restart it on each host
resp, err := c.Service.Restart(ctx, "_all", "app")
for _, r := range resp.Data.Results {
    fmt.Printf("%s changed=%v\n", r.Hostname, r.Changed)
}
```

## Example

Refer to
[`examples/sdk/client/service.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/service.go)
for a complete working example.

## Permissions

| Operation                                                     | Permission      |
| ------------------------------------------------------------- | --------------- |
| List, Get                                                     | `service:read`  |
| Create, Update, Delete, Start, Stop, Restart, Enable, Disable | `service:write` |
