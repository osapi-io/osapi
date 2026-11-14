---
sidebar_position: 1
---

# Node management

OSAPI can query and update system-level information on managed nodes. All node
operations run through the [job system](job-system.md), so the API server never
needs direct access to the host.

## Agent vs. node

OSAPI separates agent fleet discovery from node system queries:

- **Agent** commands (`agent list`, `agent get`) read the NATS KV heartbeat
  registry directly. They show the agents that are online, their labels, a small
  set of metrics and the [node conditions](agent-lifecycle.md) of the last
  heartbeat. They make no jobs. Agents also have typed **system facts**
  (architecture, kernel version, FQDN, CPU count, network interfaces, service
  manager, package manager). Providers collect them every 60 seconds, and the
  agent stores them in a separate `agent-facts` KV bucket with a 5-minute TTL.
  The API puts the registry data and the facts together in one `AgentInfo`
  response. You can [drain](agent-lifecycle.md#agent-drain) an agent for
  maintenance. The process continues to run.
- **Node** commands (`node hostname`, `node status`) send jobs to agents. The
  agents run system commands and return detailed results, for example the disk
  usage and all of the memory data.

## What it manages

| Resource                        | Description                                                     |
| ------------------------------- | --------------------------------------------------------------- |
| Hostname                        | Get or update the system hostname                               |
| Status                          | Uptime, OS name and version, kernel, platform info              |
| Disk                            | Per-mount usage (total, used, free, percent)                    |
| Memory                          | RAM and swap usage (total, used, free, percent)                 |
| Load                            | 1-, 5-, and 15-minute load averages                             |
| [System Facts](system-facts.md) | Architecture, kernel, FQDN, CPUs, NICs, routes, service/pkg mgr |

## How it works

Node queries are jobs. The CLI sends a job to the API server. The API server
publishes it to NATS. An agent receives it, reads the requested system
information and writes the result back to NATS KV. The CLI asks for the result
and shows it.

### When part of a status cannot be read

`node status` puts together six separate reads: hostname, OS information,
uptime, disks, memory and load. When one read fails, the others still return.
The response names the failure, and does not show a zero. A zero value looks the
same as a host with nothing to report.

When a read fails, the status has `partial: true` and a `field_errors` object.
Its keys are the fields that the failed reads did not fill:

```json
{
  "hostname": "web-01",
  "uptime": "4 hours, 1 minute",
  "partial": true,
  "field_errors": {
    "memory_stats": "open /proc/meminfo: permission denied"
  }
}
```

The `status` of the row stays `ok`, because the operation ran and the agent
answered. The CLI prints an **Unavailable** section with each field and the
reason that the agent could not read it.

The network and command operations are also under node. Refer to
[Network Management](network-management.md) and
[Command Execution](command-execution.md) for those domains.

You can target a specific host, broadcast to all hosts, or route by label. Refer
to [Node CLI Reference](../usage/cli/client/node/node.mdx) for job-based
commands and [Agent CLI Reference](../usage/cli/client/agent/agent.mdx) for
registry-based fleet discovery, or the
[API Reference](/gen/api/node-management-api-node-operations) for the REST
endpoints.

## Configuration

Node management uses the general job system. It needs no configuration of its
own. Refer to [Configuration](../usage/configuration.md) for NATS, agent, and
authentication settings.

## Permissions

Node read endpoints require `node:read`. Hostname update requires `node:write`.
Agent fleet discovery endpoints require `agent:read`. The built-in `admin` and
`write` roles include `node:read` and `node:write`. The `read` role includes
`node:read` only.

## Related

- [Agent CLI Reference](../usage/cli/client/agent/agent.mdx): agent fleet
  commands
- [Node CLI Reference](../usage/cli/client/node/node.mdx): node job commands
- [System Facts](system-facts.md): fact collection and `@fact.*` references
- [API Reference](/gen/api/node-management-api-node-operations): REST API
  documentation
- [Job System](job-system.md): how async job processing works
- [Architecture](../architecture/architecture.md): system design overview
