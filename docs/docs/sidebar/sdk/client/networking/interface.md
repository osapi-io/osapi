---
sidebar_position: 3
---

# Interface

Read the network interfaces of a host, and manage the Netplan config that osapi
has of one.

## Methods

| Method                            | Description                         |
| --------------------------------- | ----------------------------------- |
| `List(ctx, target)`               | List all interfaces                 |
| `Get(ctx, target, name)`          | Get a specific interface            |
| `Create(ctx, target, name, opts)` | Create a new interface config       |
| `Update(ctx, target, name, opts)` | Update an existing interface config |
| `Delete(ctx, target, name)`       | Delete an interface config          |

## Request Types

| Type                  | Fields                                                                  |
| --------------------- | ----------------------------------------------------------------------- |
| `InterfaceConfigOpts` | DHCP4, DHCP6, Addresses, Gateway4, Gateway6, MTU, MACAddress, WakeOnLAN |

## Result Types

### InterfaceListResult (List)

| Field        | Type              | Description                     |
| ------------ | ----------------- | ------------------------------- |
| `Hostname`   | `string`          | Agent hostname                  |
| `Status`     | `string`          | Result status (`ok`, `skipped`) |
| `Interfaces` | `[]InterfaceInfo` | Each interface but the loopback |
| `Error`      | `string`          | Error message (if any)          |

### InterfaceInfo

| Field       | Type       | Description                                 |
| ----------- | ---------- | ------------------------------------------- |
| `Name`      | `string`   | Interface name                              |
| `State`     | `string`   | Interface state (up, down)                  |
| `MAC`       | `string`   | Hardware MAC address                        |
| `MTU`       | `int`      | Maximum transmission unit                   |
| `IPv4`      | `string`   | First IPv4 address that is not link-local   |
| `IPv6`      | `string`   | First IPv6 address that is not link-local   |
| `Addresses` | `[]string` | Each address, in CIDR notation              |
| `DHCP4`     | `bool`     | Whether DHCP gave the interface its address |
| `Primary`   | `bool`     | Whether the interface has the default route |

### InterfaceGetResult (Get)

| Field       | Type             | Description                     |
| ----------- | ---------------- | ------------------------------- |
| `Hostname`  | `string`         | Agent hostname                  |
| `Status`    | `string`         | Result status (`ok`, `skipped`) |
| `Interface` | `*InterfaceInfo` | Interface configuration         |
| `Error`     | `string`         | Error message (if any)          |

### InterfaceMutationResult (Create, Update, Delete)

| Field      | Type     | Description                     |
| ---------- | -------- | ------------------------------- |
| `Hostname` | `string` | Agent hostname                  |
| `Status`   | `string` | Result status (`ok`, `skipped`) |
| `Name`     | `string` | Interface name                  |
| `Changed`  | `bool`   | Whether a change was made       |
| `Error`    | `string` | Error message (if any)          |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all interfaces
resp, err := c.Interface.List(ctx, "web-01")
for _, r := range resp.Data.Results {
    for _, iface := range r.Interfaces {
        fmt.Printf("%s: state=%s\n",
            iface.Name, iface.State)
    }
}

// Get a specific interface
resp, err := c.Interface.Get(ctx, "web-01", "eth0")
iface := resp.Data.First().Interface
fmt.Printf("Addresses: %v\n", iface.Addresses)

// Create a static interface config
dhcp4 := false
resp, err := c.Interface.Create(ctx, "web-01", "eth0",
    client.InterfaceConfigOpts{
        DHCP4:     &dhcp4,
        Addresses: []string{"192.168.1.100/24"},
        Gateway4:  "192.168.1.1",
    })
fmt.Printf("changed=%v\n", resp.Data.First().Changed)

// Update an existing config
resp, err := c.Interface.Update(ctx, "web-01", "eth0",
    client.InterfaceConfigOpts{
        Addresses: []string{"192.168.1.200/24"},
        Gateway4:  "192.168.1.1",
    })

// Delete an interface config
resp, err := c.Interface.Delete(ctx, "web-01", "eth0")
fmt.Printf("changed=%v\n", resp.Data.First().Changed)

// Broadcast interface create to all hosts
dhcp4 := true
resp, err := c.Interface.Create(ctx, "_all", "eth1",
    client.InterfaceConfigOpts{DHCP4: &dhcp4})
for _, r := range resp.Data.Results {
    fmt.Printf("%s changed=%v\n", r.Hostname, r.Changed)
}
```

## Example

Refer to
[`examples/sdk/client/interface.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/interface.go)
for a complete working example.

## Permissions

| Operation              | Permission        |
| ---------------------- | ----------------- |
| List, Get              | `interface:read`  |
| Create, Update, Delete | `interface:write` |

List and Get run on every platform. Create, Update and Delete run on the Debian
OS family (Ubuntu, Debian, Raspbian). On other platforms (Darwin, generic
Linux), and in a container, they return `status: skipped`. A host that refuses a
write, such as for an interface that a Netplan match defines, returns
`status: failed` with the reason in `Error`.
