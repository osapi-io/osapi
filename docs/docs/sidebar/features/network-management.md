---
sidebar_position: 2
---

# Network management

osapi can query and update the network configuration of managed hosts. Network
operations use the [job system](job-system.md). Thus the API server needs no
privileges, and the agents make the changes.

## What it manages

| Resource  | Operations           | Description                                  |
| --------- | -------------------- | -------------------------------------------- |
| DNS       | Read, Update, Delete | Nameservers and search domains per interface |
| Ping      | Read                 | ICMP connectivity check to a target host     |
| Interface | Full CRUD            | Each interface, and osapi's Netplan config   |
| Route     | Full CRUD            | Netplan static route configuration           |

For interface and route management details, see
[Network Interface Management](network-interface-management.md).

## How it works

**DNS.** DNS is the [dns behavior](../behaviors/dns/dns.md). A query reads the
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

**Ping.** Sends ICMP echo requests to a target host and reports the results.

On Linux, the agent sends a ping on an ICMP datagram socket, also when it runs
as root. The kernel permits this only for the groups in
`net.ipv4.ping_group_range`. If the group of the agent is not in the range, the
ping fails. On Debian, Ubuntu and Fedora, systemd permits each group; the
default of the kernel permits none.
[Agent hardening](agent-hardening.md#kernel-settings) tells more. To permit each
group:

```bash
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

The agent sends three echo requests and waits up to 4 seconds for the replies. A
reply that does not come back in that time is loss, in the `LOSS` column. It is
not a failure of the ping. When the job ends, for example because it is
cancelled, the ping stops.

If each ping from a host times out, but other traffic from that host works,
something between the host and the target drops ICMP. An example is a VM behind
the NAT of a Mac with the stealth mode of the macOS firewall on: the Mac drops
the replies to the pings of the VM.

Refer to [CLI Reference](../usage/cli/client/node/network/network.mdx) for usage
and examples, or the
[API Reference](/gen/api/network-management-api-network-operations) for the REST
endpoints.

## Configuration

Network management uses the general job system. It needs no configuration of its
own. Refer to [Configuration](../usage/configuration.md) for NATS, agent, and
authentication settings.

## Permissions

| Operation        | Permission      |
| ---------------- | --------------- |
| DNS get          | `network:read`  |
| DNS update       | `network:write` |
| DNS delete       | `network:write` |
| Interface list   | `network:read`  |
| Interface get    | `network:read`  |
| Interface create | `network:write` |
| Interface update | `network:write` |
| Interface delete | `network:write` |
| Ping             | `network:read`  |

The `admin` and `write` roles include both `network:read` and `network:write`.
The `read` role includes only `network:read`.

## Related

- [CLI Reference](../usage/cli/client/node/network/network.mdx): network
  commands
- [System Facts](system-facts.md): available `@fact.*` references
- [API Reference](/gen/api/network-management-api-network-operations): REST API
  documentation
- [Job System](job-system.md): how async job processing works
- [Architecture](../architecture/architecture.md): system design overview
