---
sidebar_position: 1
---

# DNS

DNS configuration query and update operations.

## Methods

| Method                                                            | Returns                                  | Permission  | Description                                      |
| ----------------------------------------------------------------- | ---------------------------------------- | ----------- | ------------------------------------------------ |
| `c.NetworkDNS.Get(ctx, hostname, interfaceName)`                  | `Response[Collection[NetworkDnsEntry]]`  | `dns:read`  | Get the DNS resolvers of an interface            |
| `c.NetworkDNS.Update(ctx, hostname, input NetworkDNSUpdateInput)` | `Response[Collection[NetworkDnsResult]]` | `dns:write` | Set the DNS resolvers of an interface            |
| `c.NetworkDNS.Delete(ctx, hostname, interfaceName)`               | `Response[Collection[NetworkDnsResult]]` | `dns:write` | Drop the DNS resolvers osapi set on an interface |

## Types

The fields of each type are the properties of its schema in the API's spec.

### NetworkDnsEntry

| Field           | Type       | Description                                                                            |
| --------------- | ---------- | -------------------------------------------------------------------------------------- |
| `Error`         | `string`   | Error message if the agent failed.                                                     |
| `Hostname`      | `string`   | Hostname of the agent that produced this result.                                       |
| `SearchDomains` | `[]string` | The search domains.                                                                    |
| `Servers`       | `[]string` | The DNS servers.                                                                       |
| `Status`        | `string`   | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### NetworkDNSUpdateInput

| Field           | Type       | Description                                                                                             |
| --------------- | ---------- | ------------------------------------------------------------------------------------------------------- |
| `InterfaceName` | `string`   | The interface_name of the path: the interface, such as eth0, or a fact such as @fact.interface.primary. |
| `OverrideDHCP`  | `bool`     | Use only these servers: drop the ones DHCP gives.                                                       |
| `SearchDomains` | `[]string` | The search domains; the interface's own when empty.                                                     |
| `Servers`       | `[]string` | The DNS servers, IPv4 or IPv6; the interface's own when empty.                                          |

### NetworkDnsResult

| Field           | Type     | Description                                                                            |
| --------------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`       | `bool`   | Whether the host changed.                                                              |
| `Error`         | `string` | Error message if the agent failed.                                                     |
| `Hostname`      | `string` | Hostname of the agent that produced this result.                                       |
| `InterfaceName` | `string` | The interface.                                                                         |
| `Status`        | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// Get DNS configuration for an interface
resp, err := c.NetworkDNS.Get(ctx, "web-01", "eth0")
for _, r := range resp.Data.Results {
    fmt.Printf("Servers: %v\n", r.Servers)
    fmt.Printf("Search:  %v\n", r.SearchDomains)
}

// Update DNS servers
resp, err := c.NetworkDNS.Update(ctx, "web-01", client.NetworkDNSUpdateInput{
    InterfaceName: "eth0",
    Servers: []string{"8.8.8.8", "8.8.4.4"},
})

// Update DNS servers and override DHCP-provided servers
resp, err = c.NetworkDNS.Update(ctx, "web-01", client.NetworkDNSUpdateInput{
    InterfaceName: "eth0",
    Servers:      []string{"1.1.1.1"},
    OverrideDHCP: true, // only use configured servers, ignore DHCP
})

// Delete managed DNS configuration
resp, err := c.NetworkDNS.Delete(ctx, "web-01", "eth0")
```

## Example

Refer to
[`examples/sdk/client/dns.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/dns.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method. Get runs on every platform. Update and Delete run on the Debian family,
and report `skipped` in a container and elsewhere.
