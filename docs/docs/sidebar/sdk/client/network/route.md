---
sidebar_position: 4
---

# Route

Static network route management via Netplan.

## Methods

| Method                                     | Description                    |
| ------------------------------------------ | ------------------------------ |
| `List(ctx, target)`                        | List all routes                |
| `Get(ctx, target, interfaceName)`          | Get routes for an interface    |
| `Create(ctx, target, interfaceName, opts)` | Create routes for an interface |
| `Update(ctx, target, interfaceName, opts)` | Update routes for an interface |
| `Delete(ctx, target, interfaceName)`       | Delete routes for an interface |

## Request Types

| Type              | Fields                  |
| ----------------- | ----------------------- |
| `RouteConfigOpts` | Routes ([]RouteItem)    |
| `RouteItem`       | To, Via, Metric         |

## Result Types

### RouteListResult (List)

| Field      | Type          | Description                     |
| ---------- | ------------- | ------------------------------- |
| `Hostname` | `string`      | Agent hostname                  |
| `Status`   | `string`      | Result status (`ok`, `skipped`) |
| `Routes`   | `[]RouteInfo` | List of routes on the host      |
| `Error`    | `string`      | Error message (if any)          |

### RouteInfo

| Field         | Type     | Description                    |
| ------------- | -------- | ------------------------------ |
| `Destination` | `string` | Destination in CIDR notation   |
| `Gateway`     | `string` | Gateway IP address             |
| `Interface`   | `string` | Network interface name         |
| `Metric`      | `int`    | Route metric (priority)        |
| `Scope`       | `string` | Route scope                    |

### RouteGetResult (Get)

| Field      | Type          | Description                     |
| ---------- | ------------- | ------------------------------- |
| `Hostname` | `string`      | Agent hostname                  |
| `Status`   | `string`      | Result status (`ok`, `skipped`) |
| `Routes`   | `[]RouteInfo` | Routes for the interface        |
| `Error`    | `string`      | Error message (if any)          |

### RouteMutationResult (Create, Update, Delete)

| Field       | Type     | Description                     |
| ----------- | -------- | ------------------------------- |
| `Hostname`  | `string` | Agent hostname                  |
| `Status`    | `string` | Result status (`ok`, `skipped`) |
| `Interface` | `string` | Interface name                  |
| `Changed`   | `bool`   | Whether a change was made       |
| `Error`     | `string` | Error message (if any)          |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all routes on the host
resp, err := c.Route.List(ctx, "web-01")
for _, r := range resp.Data.Results {
    for _, rt := range r.Routes {
        fmt.Printf("%s via %s dev %s metric %d\n",
            rt.Destination, rt.Gateway,
            rt.Interface, rt.Metric)
    }
}

// Get managed routes for a specific interface
resp, err := c.Route.Get(ctx, "web-01", "eth0")
for _, rt := range resp.Data.First().Routes {
    fmt.Printf("%s via %s\n", rt.Destination, rt.Gateway)
}

// Create static routes for an interface
metric := 100
resp, err := c.Route.Create(ctx, "web-01", "eth0",
    client.RouteConfigOpts{
        Routes: []client.RouteItem{
            {To: "10.0.0.0/8", Via: "192.168.1.1"},
            {To: "172.16.0.0/12", Via: "192.168.1.1",
             Metric: &metric},
        },
    })
fmt.Printf("changed=%v\n", resp.Data.First().Changed)

// Update routes (replace all routes for the interface)
resp, err := c.Route.Update(ctx, "web-01", "eth0",
    client.RouteConfigOpts{
        Routes: []client.RouteItem{
            {To: "10.0.0.0/8", Via: "192.168.1.254"},
        },
    })

// Delete routes for an interface
resp, err := c.Route.Delete(ctx, "web-01", "eth0")
fmt.Printf("changed=%v\n", resp.Data.First().Changed)

// Broadcast route create to all hosts
resp, err := c.Route.Create(ctx, "_all", "eth0",
    client.RouteConfigOpts{
        Routes: []client.RouteItem{
            {To: "10.0.0.0/8", Via: "192.168.1.1"},
        },
    })
for _, r := range resp.Data.Results {
    fmt.Printf("%s changed=%v\n", r.Hostname, r.Changed)
}
```

## Example

See
[`examples/sdk/client/route.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/route.go)
for a complete working example.

## Permissions

| Operation              | Permission      |
| ---------------------- | --------------- |
| List, Get              | `network:read`  |
| Create, Update, Delete | `network:write` |

Route management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
