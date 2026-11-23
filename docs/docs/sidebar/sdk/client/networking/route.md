---
sidebar_position: 4
---

# Route

Read the routes of an interface of a host, and add, change and drop the static
routes that osapi writes through Netplan, one route a call. Refer to
[Network interface management](../../../features/network-interface-management.md).

## Methods

| Method                                                                                  | Returns                                             | Permission        | Description                                 |
| --------------------------------------------------------------------------------------- | --------------------------------------------------- | ----------------- | ------------------------------------------- |
| `c.NetworkInterfaceRoute.List(ctx, hostname, name)`                                     | `Response[Collection[NetworkInterfaceRouteEntry]]`  | `interface:read`  | List the routes of an interface             |
| `c.NetworkInterfaceRoute.Create(ctx, hostname, input NetworkInterfaceRouteCreateInput)` | `Response[Collection[NetworkInterfaceRouteResult]]` | `interface:write` | Add a static route to an interface          |
| `c.NetworkInterfaceRoute.Update(ctx, hostname, input NetworkInterfaceRouteUpdateInput)` | `Response[Collection[NetworkInterfaceRouteResult]]` | `interface:write` | Change the gateway or the metric of a route |
| `c.NetworkInterfaceRoute.Delete(ctx, hostname, input NetworkInterfaceRouteDeleteInput)` | `Response[Collection[NetworkInterfaceRouteResult]]` | `interface:write` | Drop a route that osapi wrote               |

## Types

The fields of each type are the properties of its schema in the API's spec.

### NetworkInterfaceRouteEntry

| Field         | Type     | Description                                                                            |
| ------------- | -------- | -------------------------------------------------------------------------------------- |
| `Destination` | `string` | The destination in CIDR notation, or default.                                          |
| `Error`       | `string` | Error message if the agent failed.                                                     |
| `Family`      | `string` | Inet or inet6.                                                                         |
| `Gateway`     | `string` | The gateway; empty for a route on the link.                                            |
| `Hostname`    | `string` | Hostname of the agent that produced this result.                                       |
| `Interface`   | `string` | The interface the route goes out of.                                                   |
| `Metric`      | `int`    | The metric; 0 when the kernel gives none.                                              |
| `Proto`       | `string` | Where the route came from, such as kernel, dhcp or static; empty on macOS.             |
| `Scope`       | `string` | The scope, such as link; empty for a global route.                                     |
| `Status`      | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### NetworkInterfaceRouteCreateInput

| Field         | Type     | Description                                                                                   |
| ------------- | -------- | --------------------------------------------------------------------------------------------- |
| `Name`        | `string` | The name of the path: the interface, such as eth0, or a fact such as @fact.interface.primary. |
| `Destination` | `string` | Where the route goes, in CIDR notation, such as 10.0.0.0/8; not a default route.              |
| `Metric`      | `int`    | The metric, from 0 to 4294967295; none when it is 0.                                          |
| `Via`         | `string` | The gateway, an address of the destination's family, such as 192.168.1.1.                     |

### NetworkInterfaceRouteResult

| Field         | Type     | Description                                                                            |
| ------------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`     | `bool`   | Whether the host changed.                                                              |
| `Destination` | `string` | The destination of the route.                                                          |
| `Error`       | `string` | Error message if the agent failed.                                                     |
| `Hostname`    | `string` | Hostname of the agent that produced this result.                                       |
| `Status`      | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### NetworkInterfaceRouteUpdateInput

| Field         | Type     | Description                                                                                   |
| ------------- | -------- | --------------------------------------------------------------------------------------------- |
| `Name`        | `string` | The name of the path: the interface, such as eth0, or a fact such as @fact.interface.primary. |
| `Destination` | `string` | Where the route goes, in CIDR notation, such as 10.0.0.0/8; not a default route.              |
| `Metric`      | `int`    | The metric, from 0 to 4294967295; none when it is 0.                                          |
| `Via`         | `string` | The gateway, an address of the destination's family, such as 192.168.1.1.                     |

### NetworkInterfaceRouteDeleteInput

| Field         | Type     | Description                                                                                   |
| ------------- | -------- | --------------------------------------------------------------------------------------------- |
| `Name`        | `string` | The name of the path: the interface, such as eth0, or a fact such as @fact.interface.primary. |
| `Destination` | `string` | The destination of the route, in CIDR notation, such as 10.0.0.0/8.                           |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List the routes of eth0: one row per route per host
resp, err := c.NetworkInterfaceRoute.List(ctx, "web-01", "eth0")
for _, r := range resp.Data.Results {
    fmt.Printf("%s via %s metric %d\n", r.Destination, r.Gateway, r.Metric)
}

// Add a route, change its metric, then drop it
_, err = c.NetworkInterfaceRoute.Create(ctx, "web-01", client.NetworkInterfaceRouteCreateInput{
    Name: "eth1", Destination: "10.0.0.0/8", Via: "192.168.1.1",
})
_, err = c.NetworkInterfaceRoute.Update(ctx, "web-01", client.NetworkInterfaceRouteUpdateInput{
    Name: "eth1",
    Destination: "10.0.0.0/8", Via: "192.168.1.1", Metric: 100,
})
_, err = c.NetworkInterfaceRoute.Delete(ctx, "web-01", client.NetworkInterfaceRouteDeleteInput{
    Name: "eth1",
    Destination: "10.0.0.0/8",
})
```

## Example

Refer to
[`examples/sdk/client/route.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/route.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method. A route is a child of an interface. Thus it has the permissions of the
interface.

List runs on every platform. Create, Update and Delete run on the Debian OS
family (Ubuntu, Debian, Raspbian). On other platforms (Darwin, generic Linux),
and in a container, they return `status: skipped`. A default route is refused:
set the gateway of the interface with the NetworkInterface service. A host that
refuses a write returns `status: failed` with the reason in `Error`.
