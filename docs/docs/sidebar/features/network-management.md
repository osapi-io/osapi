---
sidebar_position: 2
---

# Network Management

OSAPI can query and update network configuration on managed hosts. Network
operations run through the [job system](job-system.md), keeping the API server
unprivileged while agents execute the actual changes.

## What It Manages

| Resource | Operations   | Description                                  |
| -------- | ------------ | -------------------------------------------- |
| DNS      | Read, Update | Nameservers and search domains per interface |
| Ping     | Read         | ICMP connectivity check to a target host     |

## How It Works

**DNS** -- queries read the current nameserver configuration for a network
interface. Updates modify the nameservers and search domains, applying changes
through the host's network manager.

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
- [API Reference](/gen/api/network-management-api-network-operations) -- REST
  API documentation
- [Job System](job-system.md) -- how async job processing works
- [Architecture](../architecture/architecture.md) -- system design overview
