---
sidebar_position: 1
---

# Node Management

OSAPI can query a variety of system-level information from managed nodes. All
node operations are read-only today and run through the
[job system](job-system.md), so the API server never needs direct access to the
host.

## Agent vs. Node

OSAPI separates agent fleet discovery from node system queries:

- **Agent** commands (`agent list`, `agent get`) read directly from the NATS KV
  heartbeat registry. They show which agents are online, their labels, and
  lightweight metrics from the last heartbeat. No jobs are created.
- **Node** commands (`node hostname`, `node status`) dispatch jobs to agents that
  execute system commands and return detailed results (disk usage, full memory
  breakdown, etc.).

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
API server publishes it to NATS, a agent picks it up and reads the
requested system information, then writes the result back to NATS KV. The CLI
polls for the result and displays it.

You can target a specific host, broadcast to all hosts, or route by label. See
[Node CLI Reference](../usage/cli/client/node/node.mdx) for job-based commands
and [Agent CLI Reference](../usage/cli/client/agent/agent.mdx) for registry-based
fleet discovery, or the
[API Reference](/gen/api/node-management-api-node-operations) for the REST
endpoints.

## Configuration

Node management uses the general job infrastructure. No domain-specific
configuration is required. See [Configuration](../usage/configuration.md) for
NATS, agent, and authentication settings.

## Permissions

Node job endpoints require the `node:read` permission. Agent fleet discovery
endpoints require the `agent:read` permission. The built-in `admin`, `write`,
and `read` roles all include both permissions.

## Related

- [Agent CLI Reference](../usage/cli/client/agent/agent.mdx) -- agent fleet
  commands
- [Node CLI Reference](../usage/cli/client/node/node.mdx) -- node job commands
- [API Reference](/gen/api/node-management-api-node-operations) -- REST API
  documentation
- [Job System](job-system.md) -- how async job processing works
- [Architecture](../architecture/architecture.md) -- system design overview
