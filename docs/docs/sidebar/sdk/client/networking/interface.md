---
sidebar_position: 3
---

# Interface

Read the network interfaces of a host, and manage the Netplan config that osapi
has of one.

## Methods

| Method                                                                        | Returns                                        | Permission        | Description                                        |
| ----------------------------------------------------------------------------- | ---------------------------------------------- | ----------------- | -------------------------------------------------- |
| `c.NetworkInterface.List(ctx, hostname)`                                      | `Response[Collection[NetworkInterfaceEntry]]`  | `interface:read`  | List the network interfaces of the host            |
| `c.NetworkInterface.Get(ctx, hostname, name)`                                 | `Response[Collection[NetworkInterfaceEntry]]`  | `interface:read`  | Get a network interface of the host                |
| `c.NetworkInterface.Create(ctx, hostname, input NetworkInterfaceCreateInput)` | `Response[Collection[NetworkInterfaceResult]]` | `interface:write` | Configure a network interface                      |
| `c.NetworkInterface.Update(ctx, hostname, input NetworkInterfaceUpdateInput)` | `Response[Collection[NetworkInterfaceResult]]` | `interface:write` | Change the config osapi has of a network interface |
| `c.NetworkInterface.Delete(ctx, hostname, name)`                              | `Response[Collection[NetworkInterfaceResult]]` | `interface:write` | Drop the config osapi has of a network interface   |

## Types

The fields of each type are the properties of its schema in the API's spec.

### NetworkInterfaceEntry

| Field       | Type       | Description                                                                            |
| ----------- | ---------- | -------------------------------------------------------------------------------------- |
| `Addresses` | `[]string` | Each of its addresses, in CIDR notation.                                               |
| `DHCP4`     | `bool`     | Whether DHCP gave it an IPv4 address: a route of it has proto dhcp.                    |
| `Error`     | `string`   | Error message if the agent failed.                                                     |
| `Hostname`  | `string`   | Hostname of the agent that produced this result.                                       |
| `IPv4`      | `string`   | Its first IPv4 address that is not link-local.                                         |
| `IPv6`      | `string`   | Its first IPv6 address that is not link-local.                                         |
| `MAC`       | `string`   | Its MAC address; empty for one without, such as lo.                                    |
| `MTU`       | `int`      | Its MTU.                                                                               |
| `Name`      | `string`   | The interface.                                                                         |
| `Primary`   | `bool`     | Whether it has the default route.                                                      |
| `State`     | `string`   | Up or down.                                                                            |
| `Status`    | `string`   | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### NetworkInterfaceCreateInput

| Field        | Type       | Description                                                                                   |
| ------------ | ---------- | --------------------------------------------------------------------------------------------- |
| `Name`       | `string`   | The name of the path: the interface, such as eth0, or a fact such as @fact.interface.primary. |
| `Addresses`  | `[]string` | The static addresses, in CIDR notation, such as 10.0.0.5/24.                                  |
| `DHCP4`      | `bool`     | Take an IPv4 address from DHCP.                                                               |
| `DHCP6`      | `bool`     | Take an IPv6 address from DHCP.                                                               |
| `Gateway4`   | `string`   | The IPv4 default gateway.                                                                     |
| `Gateway6`   | `string`   | The IPv6 default gateway.                                                                     |
| `MACAddress` | `string`   | The MAC address to set.                                                                       |
| `MTU`        | `int`      | The MTU, from 68 to 9000.                                                                     |
| `Wakeonlan`  | `bool`     | Turn on wake-on-LAN.                                                                          |

### NetworkInterfaceResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Name`     | `string` | The interface.                                                                         |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### NetworkInterfaceUpdateInput

| Field        | Type       | Description                                                                                   |
| ------------ | ---------- | --------------------------------------------------------------------------------------------- |
| `Name`       | `string`   | The name of the path: the interface, such as eth0, or a fact such as @fact.interface.primary. |
| `Addresses`  | `[]string` | The static addresses, in CIDR notation, such as 10.0.0.5/24.                                  |
| `DHCP4`      | `bool`     | Take an IPv4 address from DHCP.                                                               |
| `DHCP6`      | `bool`     | Take an IPv6 address from DHCP.                                                               |
| `Gateway4`   | `string`   | The IPv4 default gateway.                                                                     |
| `Gateway6`   | `string`   | The IPv6 default gateway.                                                                     |
| `MACAddress` | `string`   | The MAC address to set.                                                                       |
| `MTU`        | `int`      | The MTU, from 68 to 9000.                                                                     |
| `Wakeonlan`  | `bool`     | Turn on wake-on-LAN.                                                                          |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all interfaces
resp, err := c.NetworkInterface.List(ctx, "web-01")
for _, r := range resp.Data.Results {
    fmt.Printf("%s: state=%s\n", r.Name, r.State)
}

// Get a specific interface
resp, err := c.NetworkInterface.Get(ctx, "web-01", "eth0")
if iface, ok := resp.Data.First(); ok {
    fmt.Printf("Addresses: %v\n", iface.Addresses)
}

// Create a static interface config
resp, err := c.NetworkInterface.Create(ctx, "web-01", client.NetworkInterfaceCreateInput{
    Name: "eth0",
    Addresses: []string{"192.168.1.100/24"},
    Gateway4:  "192.168.1.1",
})

// Update an existing config
resp, err := c.NetworkInterface.Update(ctx, "web-01", client.NetworkInterfaceUpdateInput{
    Name: "eth0",
    Addresses: []string{"192.168.1.200/24"},
    Gateway4:  "192.168.1.1",
})

// Delete an interface config
resp, err := c.NetworkInterface.Delete(ctx, "web-01", "eth0")

// Broadcast interface create to all hosts
resp, err := c.NetworkInterface.Create(ctx, "_all", client.NetworkInterfaceCreateInput{
    Name: "eth1",
    DHCP4: true,
})
for _, r := range resp.Data.Results {
    fmt.Printf("%s changed=%v\n", r.Hostname, r.Changed)
}
```

## Example

Refer to
[`examples/sdk/client/interface.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/interface.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.

List and Get run on every platform. Create, Update and Delete run on the Debian
OS family (Ubuntu, Debian, Raspbian). On other platforms (Darwin, generic
Linux), and in a container, they return `status: skipped`. A host that refuses a
write, such as for an interface that a Netplan match defines, returns
`status: failed` with the reason in `Error`.
