---
sidebar_position: 5
---

# Service

Read the systemd services of target hosts. Deploy, remove, start, stop, restart,
enable and disable the services that osapi deploys. The unit file of each is an
object that `File.Upload` stores one time. Refer to
[Service management](../../../features/service-management.md).

## Methods

| Method                                                      | Returns                               | Permission        | Description                                          |
| ----------------------------------------------------------- | ------------------------------------- | ----------------- | ---------------------------------------------------- |
| `c.Service.List(ctx, hostname)`                             | `Response[Collection[ServiceEntry]]`  | `service:read`    | List the services of the host                        |
| `c.Service.Create(ctx, hostname, input ServiceCreateInput)` | `Response[Collection[ServiceResult]]` | `service:write`   | Deploy a service                                     |
| `c.Service.Get(ctx, hostname, name)`                        | `Response[Collection[ServiceEntry]]`  | `service:read`    | Get one service                                      |
| `c.Service.Update(ctx, hostname, input ServiceUpdateInput)` | `Response[Collection[ServiceResult]]` | `service:write`   | Deploy a new unit file for a service                 |
| `c.Service.Delete(ctx, hostname, name)`                     | `Response[Collection[ServiceResult]]` | `service:write`   | Remove a service that osapi deployed                 |
| `c.Service.Disable(ctx, hostname, name)`                    | `Response[Collection[ServiceResult]]` | `service:execute` | Make a service that osapi deployed not start at boot |
| `c.Service.Enable(ctx, hostname, name)`                     | `Response[Collection[ServiceResult]]` | `service:execute` | Make a service that osapi deployed start at boot     |
| `c.Service.Restart(ctx, hostname, name)`                    | `Response[Collection[ServiceResult]]` | `service:execute` | Restart a service that osapi deployed                |
| `c.Service.Start(ctx, hostname, name)`                      | `Response[Collection[ServiceResult]]` | `service:execute` | Start a service that osapi deployed                  |
| `c.Service.Stop(ctx, hostname, name)`                       | `Response[Collection[ServiceResult]]` | `service:execute` | Stop a service that osapi deployed                   |

## Types

The fields of each type are the properties of its schema in the API's spec.

### ServiceEntry

| Field         | Type     | Description                                                                            |
| ------------- | -------- | -------------------------------------------------------------------------------------- |
| `Description` | `string` | The Description= of its unit.                                                          |
| `Enabled`     | `bool`   | Whether it starts at boot: the state of its unit file.                                 |
| `Error`       | `string` | Error message if the agent failed.                                                     |
| `Hostname`    | `string` | Hostname of the agent that produced this result.                                       |
| `Managed`     | `bool`   | Whether osapi deployed it, so its writes and actions act on it.                        |
| `Name`        | `string` | The service, without .service.                                                         |
| `State`       | `string` | What systemd says it does now, such as running, exited or dead.                        |
| `Status`      | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### ServiceCreateInput

| Field    | Type     | Description                                                         |
| -------- | -------- | ------------------------------------------------------------------- |
| `Name`   | `string` | The service, such as app for osapi-app.service.                     |
| `Object` | `string` | The object that holds the unit file, from osapi client file upload. |

### ServiceResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Name`     | `string` | The service.                                                                           |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### ServiceUpdateInput

| Field    | Type     | Description                                                           |
| -------- | -------- | --------------------------------------------------------------------- |
| `Name`   | `string` | The name of the path: the service, such as app for osapi-app.service. |
| `Object` | `string` | The object that holds the unit file, from osapi client file upload.   |

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
    client.ServiceCreateInput{Name: "app", Object: "app.service"})
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

The Permission column of [Methods](#methods) gives the permission of each
method.
