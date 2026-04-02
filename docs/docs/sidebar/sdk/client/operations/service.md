---
sidebar_position: 5
---

# Service

Systemd service management on target hosts. List, inspect, start, stop, restart,
enable, disable services, and manage unit files via the Object Store.

## Methods

| Method                              | Description                          |
| ----------------------------------- | ------------------------------------ |
| `List(ctx, hostname)`               | List all services                    |
| `Get(ctx, hostname, name)`          | Get details for a specific service   |
| `Start(ctx, hostname, name)`        | Start a service                      |
| `Stop(ctx, hostname, name)`         | Stop a service                       |
| `Restart(ctx, hostname, name)`      | Restart a service                    |
| `Enable(ctx, hostname, name)`       | Enable a service to start on boot    |
| `Disable(ctx, hostname, name)`      | Disable a service from boot          |
| `Create(ctx, hostname, opts)`       | Deploy a new unit file               |
| `Update(ctx, hostname, name, opts)` | Update an existing unit file         |
| `Delete(ctx, hostname, name)`       | Delete a unit file                   |

## Request Types

| Type                | Fields                             |
| ------------------- | ---------------------------------- |
| `ServiceCreateOpts` | Name (required), Object (required) |
| `ServiceUpdateOpts` | Object (required)                  |

## Result Types

### ServiceInfoResult (List)

| Field      | Type            | Description                       |
| ---------- | --------------- | --------------------------------- |
| `Hostname` | `string`        | Agent hostname                    |
| `Status`   | `string`        | Result status (`ok`, `skipped`)   |
| `Services` | `[]ServiceInfo` | List of services on the host      |
| `Error`    | `string`        | Error message (if any)            |

### ServiceInfo

| Field         | Type     | Description                       |
| ------------- | -------- | --------------------------------- |
| `Name`        | `string` | Service unit name                 |
| `Status`      | `string` | Active status (active, inactive)  |
| `Enabled`     | `bool`   | Whether the service starts on boot |
| `Description` | `string` | Service description               |
| `PID`         | `int`    | Process ID (if running)           |

### ServiceGetResult (Get)

| Field      | Type           | Description                       |
| ---------- | -------------- | --------------------------------- |
| `Hostname` | `string`       | Agent hostname                    |
| `Status`   | `string`       | Result status (`ok`, `skipped`)   |
| `Service`  | `*ServiceInfo` | Service details                   |
| `Error`    | `string`       | Error message (if any)            |

### ServiceMutationResult (mutations and actions)

| Field      | Type     | Description                       |
| ---------- | -------- | --------------------------------- |
| `Hostname` | `string` | Agent hostname                    |
| `Status`   | `string` | Result status (`ok`, `skipped`)   |
| `Name`     | `string` | Service unit name                 |
| `Changed`  | `bool`   | Whether a change was made         |
| `Error`    | `string` | Error message (if any)            |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all services
resp, err := c.Service.List(ctx, "web-01")
for _, r := range resp.Data.Results {
    for _, svc := range r.Services {
        fmt.Printf("%s status=%s enabled=%v\n",
            svc.Name, svc.Status, svc.Enabled)
    }
}

// Get a specific service
resp, err := c.Service.Get(ctx, "web-01", "nginx.service")
svc := resp.Data.First().Service
fmt.Printf("%s pid=%d\n", svc.Name, svc.PID)

// Start a service
resp, err := c.Service.Start(ctx, "web-01", "nginx.service")
fmt.Printf("changed=%v\n", resp.Data.First().Changed)

// Restart a service across all hosts
resp, err := c.Service.Restart(ctx, "_all", "nginx.service")
for _, r := range resp.Data.Results {
    fmt.Printf("%s changed=%v\n", r.Hostname, r.Changed)
}

// Deploy a new unit file
resp, err := c.Service.Create(ctx, "web-01",
    client.ServiceCreateOpts{
        Name:   "myapp.service",
        Object: "myapp-unit",
    })

// Update an existing unit file
resp, err := c.Service.Update(ctx, "web-01", "myapp.service",
    client.ServiceUpdateOpts{
        Object: "myapp-unit-v2",
    })

// Delete a unit file
resp, err := c.Service.Delete(ctx, "web-01", "myapp.service")
```

## Example

See
[`examples/sdk/client/service.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/service.go)
for a complete working example.

## Permissions

| Operation                                                     | Permission      |
| ------------------------------------------------------------- | --------------- |
| List, Get                                                     | `service:read`  |
| Start, Stop, Restart, Enable, Disable, Create, Update, Delete | `service:write` |

Service management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
