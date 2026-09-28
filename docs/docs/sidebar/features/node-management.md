---
sidebar_position: 1
---

# Node Management

OSAPI can query and update system-level information on managed nodes. All node
operations run through the [job system](job-system.md), so the API server never
needs direct access to the host.

## Agent vs. Node

OSAPI separates agent fleet discovery from node system queries:

- **Agent** commands (`agent list`, `agent get`) read directly from the NATS KV
  heartbeat registry. They show which agents are online, their labels,
  lightweight metrics, and [node conditions](agent-lifecycle.md) from the last
  heartbeat. No jobs are created. Agents also expose typed **system facts**
  (architecture, kernel version, FQDN, CPU count, network interfaces, service
  manager, package manager) gathered every 60 seconds via providers and stored
  in a separate `agent-facts` KV bucket with a 5-minute TTL. The API merges
  registry and facts data into a single `AgentInfo` response. Agents can be
  [drained](agent-lifecycle.md#agent-drain) for maintenance without stopping the
  process.
- **Node** commands (`node hostname`, `node status`) dispatch jobs to agents
  that execute system commands and return detailed results (disk usage, full
  memory breakdown, etc.).

## What It Manages

| Resource                        | Description                                                     |
| ------------------------------- | --------------------------------------------------------------- |
| Hostname                        | Get or update the system hostname                               |
| Status                          | Uptime, OS name and version, kernel, platform info              |
| Disk                            | Per-mount usage (total, used, free, percent)                    |
| Memory                          | RAM and swap usage (total, used, free, percent)                 |
| Load                            | 1-, 5-, and 15-minute load averages                             |
| [System Facts](system-facts.md) | Architecture, kernel, FQDN, CPUs, NICs, routes, service/pkg mgr |

## How It Works

Node queries are submitted as jobs. The CLI posts a job to the API server, the
API server publishes it to NATS, an agent picks it up and reads the requested
system information, then writes the result back to NATS KV. The CLI polls for
the result and displays it.

### When part of a status cannot be read

`node status` is assembled from six independent reads: hostname, OS info,
uptime, disks, memory and load. One of them failing is not a reason to answer
nothing, so the others are still returned — but the failure is named rather than
left as a zero, because a zero-valued field is otherwise indistinguishable from
a host with nothing to report.

A status where a read failed carries `partial: true` and a `field_errors` object
keyed by the field each failed read would have filled:

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

The row's `status` stays `ok`: the operation ran, and the agent answered. The
CLI prints an **Unavailable** section listing each field and why it could not be
read.

Network and command operations are also nested under the node — see
[Network Management](network-management.md) and
[Command Execution](command-execution.md) for those domains.

You can target a specific host, broadcast to all hosts, or route by label. See
[Node CLI Reference](../usage/cli/client/node/node.mdx) for job-based commands
and [Agent CLI Reference](../usage/cli/client/agent/agent.mdx) for
registry-based fleet discovery, or the
[API Reference](/gen/api/node-management-api-node-operations) for the REST
endpoints.

## Configuration

Node management uses the general job infrastructure. No domain-specific
configuration is required. See [Configuration](../usage/configuration.md) for
NATS, agent, and authentication settings.

## Permissions

Node read endpoints require `node:read`. Hostname update requires `node:write`.
Agent fleet discovery endpoints require `agent:read`. The built-in `admin` and
`write` roles include `node:read` and `node:write`. The `read` role includes
`node:read` only.

## Related

- [Agent CLI Reference](../usage/cli/client/agent/agent.mdx) -- agent fleet
  commands
- [Node CLI Reference](../usage/cli/client/node/node.mdx) -- node job commands
- [System Facts](system-facts.md) -- fact collection and `@fact.*` references
- [API Reference](/gen/api/node-management-api-node-operations) -- REST API
  documentation
- [Job System](job-system.md) -- how async job processing works
- [Architecture](../architecture/architecture.md) -- system design overview
