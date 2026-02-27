---
sidebar_position: 1
---

# Node Management

OSAPI can query a variety of system-level information from managed nodes. All
node operations are read-only today and run through the
[job system](job-system.md), so the API server never needs direct access to the
host.

## What It Manages

| Resource | Description                                        |
| -------- | -------------------------------------------------- |
| Hostname | System hostname                                    |
| Status   | Uptime, OS name and version, kernel, platform info |
| Disk     | Per-mount usage (total, used, free, percent)       |
| Memory   | RAM and swap usage (total, used, free, percent)    |
| Load     | 1-, 5-, and 15-minute load averages                |

## How It Works

Node queries are submitted as jobs. The CLI posts a job to the API server, the
API server publishes it to NATS, a node agent picks it up and reads the
requested system information, then writes the result back to NATS KV. The CLI
polls for the result and displays it.

You can target a specific host, broadcast to all hosts, or route by label. See
[CLI Reference](../usage/cli/client/node/node.mdx) for usage and examples, or
the [API Reference](/gen/api/node-management-api-node-operations) for the REST
endpoints.

## Configuration

Node management uses the general job infrastructure. No domain-specific
configuration is required. See [Configuration](../usage/configuration.md) for
NATS, node agent, and authentication settings.

## Permissions

All node endpoints require the `node:read` permission. The built-in `admin`,
`write`, and `read` roles all include this permission.

## Related

- [CLI Reference](../usage/cli/client/node/node.mdx) -- node commands
- [API Reference](/gen/api/node-management-api-node-operations) -- REST API
  documentation
- [Job System](job-system.md) -- how async job processing works
- [Architecture](../architecture/architecture.md) -- system design overview
