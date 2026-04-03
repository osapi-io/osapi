---
sidebar_position: 2
---

# Network Management

OSAPI can query and update network configuration on managed hosts. Network
operations run through the [job system](job-system.md), keeping the API server
unprivileged while agents execute the actual changes.

## What It Manages

| Resource  | Operations          | Description                                  |
| --------- | ------------------- | -------------------------------------------- |
| DNS       | Read, Update, Delete| Nameservers and search domains per interface |
| Ping      | Read                | ICMP connectivity check to a target host     |
| Interface | Full CRUD           | Netplan interface configuration              |
| Route     | Full CRUD           | Netplan static route configuration           |

For interface and route management details, see
[Network Interface Management](network-interface-management.md).

## How It Works

**DNS** -- queries read the current nameserver configuration for a network
interface via `resolvectl`. Updates generate a persistent Netplan configuration
file (`/etc/netplan/osapi-dns.yaml`) targeting the primary interface, validate
with `netplan generate`, and apply with `netplan apply`. This ensures DNS
changes survive reboots. The `--interface-name` parameter supports
[fact references](system-facts.md) — use `@fact.interface.primary` to
automatically target the default route interface.

**Ping** -- sends ICMP echo requests to a target host and reports the results.

See [CLI Reference](../usage/cli/client/node/network/network.mdx) for usage and
examples, or the
[API Reference](/gen/api/network-management-api-network-operations) for the REST
endpoints.

## Configuration

Network management uses the general job infrastructure. No domain-specific
configuration is required. See [Configuration](../usage/configuration.md) for
NATS, agent, and authentication settings.

## Permissions

| Operation  | Permission      |
| ---------- | --------------- |
| DNS get    | `network:read`  |
| DNS update | `network:write` |
| Ping       | `network:read`  |

The `admin` and `write` roles include both `network:read` and `network:write`.
The `read` role includes only `network:read`.

## Related

- [CLI Reference](../usage/cli/client/node/network/network.mdx) -- network
  commands
- [System Facts](system-facts.md) -- available `@fact.*` references
- [API Reference](/gen/api/network-management-api-network-operations) -- REST
  API documentation
- [Job System](job-system.md) -- how async job processing works
- [Architecture](../architecture/architecture.md) -- system design overview
