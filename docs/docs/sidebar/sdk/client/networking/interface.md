---
sidebar_position: 3
---

# Interface

Network interface configuration management via Netplan.

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
| `Interfaces` | `[]InterfaceInfo` | List of interfaces on the host  |
| `Error`      | `string`          | Error message (if any)          |

### InterfaceInfo

| Field        | Type       | Description                    |
| ------------ | ---------- | ------------------------------ |
| `Name`       | `string`   | Interface name                 |
| `DHCP4`      | `bool`     | Whether DHCPv4 is enabled      |
| `DHCP6`      | `bool`     | Whether DHCPv6 is enabled      |
| `Addresses`  | `[]string` | IP addresses in CIDR notation  |
| `Gateway4`   | `string`   | IPv4 gateway address           |
| `Gateway6`   | `string`   | IPv6 gateway address           |
| `MTU`        | `int`      | Maximum transmission unit      |
| `MACAddress` | `string`   | Hardware MAC address           |
| `WakeOnLAN`  | `bool`     | Whether Wake-on-LAN is enabled |
| `State`      | `string`   | Interface state (up, down)     |

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

See
[`examples/sdk/client/interface.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/interface.go)
for a complete working example.

## Permissions

| Operation              | Permission      |
| ---------------------- | --------------- |
| List, Get              | `network:read`  |
| Create, Update, Delete | `network:write` |

Interface management is supported on the Debian OS family (Ubuntu, Debian,
Raspbian). On unsupported platforms (Darwin, generic Linux), operations return
`status: skipped`.
