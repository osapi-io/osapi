---
sidebar_position: 4
---

# Route

Read the routes of an interface of a host, and add, change and drop the static
routes that osapi writes through Netplan, one route a call. Refer to
[Network interface management](../../../features/network-interface-management.md).

## Methods

| Method                                            | Description                                    |
| ------------------------------------------------- | ---------------------------------------------- |
| `List(ctx, target, interfaceName)`                | List the routes of an interface, one row each  |
| `Create(ctx, target, interfaceName, opts)`        | Add a route; the same route again is unchanged |
| `Update(ctx, target, interfaceName, opts)`        | Change the gateway or metric of a route        |
| `Delete(ctx, target, interfaceName, destination)` | Drop a route that osapi wrote                  |

## Request types

| Type        | Fields                                                      |
| ----------- | ----------------------------------------------------------- |
| `RouteOpts` | Destination (CIDR), Via (same family), Metric (none when 0) |

## Result types

### RouteInfo (List)

| Field         | Type     | Description                                    |
| ------------- | -------- | ---------------------------------------------- |
| `Hostname`    | `string` | Agent hostname                                 |
| `Status`      | `string` | Result status (`ok`, `skipped`, `failed`)      |
| `Destination` | `string` | Destination in CIDR notation, or `default`     |
| `Gateway`     | `string` | Gateway address; empty for a route on the link |
| `Interface`   | `string` | Network interface name                         |
| `Family`      | `string` | `inet` or `inet6`                              |
| `Metric`      | `int`    | Route metric (priority)                        |
| `Proto`       | `string` | Where the route came from, such as `dhcp`      |
| `Scope`       | `string` | Route scope, such as `link`; empty for global  |
| `Error`       | `string` | Error message (if any)                         |

### RouteMutationResult (Create, Update, Delete)

| Field         | Type     | Description                               |
| ------------- | -------- | ----------------------------------------- |
| `Hostname`    | `string` | Agent hostname                            |
| `Status`      | `string` | Result status (`ok`, `skipped`, `failed`) |
| `Destination` | `string` | The destination of the route              |
| `Changed`     | `bool`   | Whether the host changed                  |
| `Error`       | `string` | Error message (if any)                    |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List the routes of eth0: one row per route per host
resp, err := c.Route.List(ctx, "web-01", "eth0")
for _, r := range resp.Data.Results {
    fmt.Printf("%s via %s metric %d\n", r.Destination, r.Gateway, r.Metric)
}

// Add a route, change its metric, then drop it
route := client.RouteOpts{Destination: "10.0.0.0/8", Via: "192.168.1.1"}
_, err = c.Route.Create(ctx, "web-01", "eth1", route)
route.Metric = 100
_, err = c.Route.Update(ctx, "web-01", "eth1", route)
_, err = c.Route.Delete(ctx, "web-01", "eth1", "10.0.0.0/8")
```

## Example

Refer to
[`examples/sdk/client/route.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/route.go)
for a complete working example.

## Permissions

| Operation              | Permission      |
| ---------------------- | --------------- |
| List                   | `network:read`  |
| Create, Update, Delete | `network:write` |

List runs on every platform. Create, Update and Delete run on the Debian OS
family (Ubuntu, Debian, Raspbian). On other platforms (Darwin, generic Linux),
and in a container, they return `status: skipped`. A default route is refused:
set the gateway of the interface with the Interface service. A host that refuses
a write returns `status: failed` with the reason in `Error`.
