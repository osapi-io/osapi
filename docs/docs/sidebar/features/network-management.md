---
sidebar_position: 2
---

# Network management

osapi can query and update the network configuration of managed hosts. Network
operations use the [job system](job-system.md). Thus the API server needs no
privileges, and the agents make the changes.

## What it manages

| Resource  | Operations           | Description                                   |
| --------- | -------------------- | --------------------------------------------- |
| DNS       | Read, Update, Delete | Nameservers and search domains per interface  |
| Ping      | Action               | ICMP connectivity check to a target host      |
| Interface | Full CRUD            | Each interface, and osapi's Netplan config    |
| Route     | Full CRUD            | The static routes of an interface, one a call |

For interface and route management details, see
[Network Interface Management](network-interface-management.md).

## How it works

**DNS.** DNS is the
[dns behavior](../reference/behaviors/networking/dns/dns.md). A query reads the
servers and the search domains of an interface through gohai on every platform:
`resolvectl` on Linux, `resolv.conf` on a host that has no resolvers per
interface, such as a container, and `scutil` on macOS. An update writes a
Netplan file of the interface's own, `/etc/netplan/osapi-dns-<interface>.yaml`,
on the Debian family. It validates the file with `netplan generate`, removes it
again when that fails, and applies it with `netplan apply`. Thus DNS changes
stay after a reboot, and the resolvers of each interface are separate. An update
or a delete reports skipped in a container and on other platforms.

The `--interface-name` parameter accepts [fact references](system-facts.md). Use
`@fact.interface.primary` to target the default route interface automatically.
Use `--override-dhcp` to disable the DNS servers from DHCP, so that the host
uses only the configured servers. Without it, the host uses the DHCP DNS servers
and the configured servers together (the default of Netplan).

:::note IPv6 Router Advertisement DNS

`--override-dhcp` disables DNS from DHCPv4 and DHCPv6. It does **not** disable
DNS from IPv6 Router Advertisements (RA/SLAAC). If you disable RA, the host
loses IPv6: its default route, its global addresses and its prefix information.
Thus DNS servers from IPv6 RA can still show in the `resolvectl` output, with
the configured servers.

:::

**Ping.** Ping is the `ping` action of the
[connectivity behavior](../reference/behaviors/networking/connectivity/connectivity.md),
`osapi client node network connectivity ping --address <ip>`. It runs the host's
`ping`, three echo requests, and reports what came back: the packets sent and
received, the loss, and the round-trip times. A host that does not answer is
100% loss, not an error. Ping changes nothing on the host, so it reports
`changed: false`. On macOS it runs `ping6` for an IPv6 address.

On Linux, the `ping` of iputils sends on an ICMP datagram socket, also when the
agent runs as root. The kernel permits this only for the groups in
`net.ipv4.ping_group_range`. If the group of the agent is not in the range, the
ping fails. On Debian, Ubuntu and Fedora, systemd permits each group; the
default of the kernel permits none.
[Agent hardening](agent-hardening.md#kernel-settings) tells more. To permit each
group:

```bash
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

`ping` sends three packets, one a second, and waits two seconds after the last,
about four seconds in all. A reply that does not come back in that time is loss,
in the `LOSS` column. It is not a failure of the ping. When the job ends, for
example because it is cancelled, the ping stops.

If each ping from a host times out, but other traffic from that host works,
something between the host and the target drops ICMP. An example is a VM behind
the NAT of a Mac with the stealth mode of the macOS firewall on: the Mac drops
the replies to the pings of the VM.

Refer to [CLI Reference](../usage/cli/client/node/network/network.mdx) for usage
and examples, or the
[API Reference](/gen/api/connectivity-api-network-connectivity-operations) for
the REST endpoints.

## Configuration

Network management uses the general job system. It needs no configuration of its
own. Refer to [Configuration](../usage/configuration.md) for NATS, agent, and
authentication settings.

## Permissions and platforms

The generated page of each behavior gives the permission, the platforms and the
route of each method: [dns](../reference/behaviors/networking/dns/dns.md),
[interface](../reference/behaviors/networking/interface/interface.md),
[route](../reference/behaviors/networking/interface/route/route.md) and
[connectivity](../reference/behaviors/networking/connectivity/connectivity.md).
The pages come from the definitions, so they stay true when a method changes,
and this page does not copy them. [Authentication](authentication.md) tells
which permissions each role has.

## Related

- [CLI Reference](../usage/cli/client/node/network/network.mdx): network
  commands
- [System Facts](system-facts.md): available `@fact.*` references
- API Reference: [DNS](/gen/api/dns-api-network-dns-operations),
  [interfaces](/gen/api/interface-api-network-interface-operations),
  [routes](/gen/api/route-api-route-operations) and
  [connectivity](/gen/api/connectivity-api-network-connectivity-operations)
- [Job System](job-system.md): how async job processing works
- [Architecture](../architecture/architecture.md): system design overview
