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

| Resource                                                | Description                                                     |
| ------------------------------------------------------- | --------------------------------------------------------------- |
| [Hostname](../reference/behaviors/hostname/hostname.md) | Get the hostname, or set it on Debian                           |
| Status                                                  | Uptime, OS name and version, kernel, platform info              |
| [Uptime](../reference/behaviors/uptime/uptime.md)       | The time since the host started                                 |
| [OS](../reference/behaviors/os/os.md)                   | The distribution and its version                                |
| [Disk](../reference/behaviors/disk/disk.md)             | The size, used and free space of each local filesystem          |
| [Memory](../reference/behaviors/memory/memory.md)       | The total, used and free memory                                 |
| [Load](../reference/behaviors/load/load.md)             | The 1, 5 and 15 minute load averages                            |
| [System Facts](system-facts.md)                         | Architecture, kernel, FQDN, CPUs, NICs, routes, service/pkg mgr |

## How it works

Node queries are jobs. The CLI sends a job to the API server. The API server
publishes it to NATS. An agent receives it, reads the requested system
information and writes the result back to NATS KV. The CLI asks for the result
and shows it.

### When part of a status cannot be read

`node status` is the [status behavior](../reference/behaviors/status/status.md).
It runs six separate reads: hostname, uptime, OS, load, memory and disks. When
one read fails, the others still return. The response names the failure in
`errors`, and does not show a zero. A zero value looks the same as a host with
nothing to report.

Each item of `errors` is the name of the read and its error:

```json
{
  "hostname": "web-01",
  "status": "ok",
  "name": "web-01",
  "uptime": { "uptime": "4 hours, 1 minute", "seconds": 14460 },
  "errors": ["memory: open /proc/meminfo: permission denied"]
}
```

The `status` of the row stays `ok`, because the operation ran and the agent
answered.

The network and command operations are also under node. Refer to
[Network Management](network-management.md) and
[Command Execution](command-execution.md) for those domains.

You can target a specific host, broadcast to all hosts, or route by label. Refer
to [Node CLI Reference](../usage/cli/client/node/node.mdx) for job-based
commands and [Agent CLI Reference](../usage/cli/client/agent/agent.mdx) for
registry-based fleet discovery, or the
[API Reference](/gen/api/status-api-status-operations) for the REST endpoints.

## Configuration

Node management uses the general job system. It needs no configuration of its
own. Refer to [Configuration](../usage/configuration.md) for NATS, agent, and
authentication settings.

## Permissions and platforms

The generated page of each behavior gives the permission, the platforms and the
route of each method: [status](../reference/behaviors/status/status.md),
[hostname](../reference/behaviors/hostname/hostname.md),
[uptime](../reference/behaviors/uptime/uptime.md),
[os](../reference/behaviors/os/os.md),
[disk](../reference/behaviors/disk/disk.md),
[memory](../reference/behaviors/memory/memory.md),
[load](../reference/behaviors/load/load.md). The pages come from the
definitions, so they stay true when a method changes, and this page does not
copy them. [Authentication](authentication.md) tells which permissions each role
has.

Agent fleet discovery is not a behavior: its endpoints need `agent:read`.

## Related

- [Agent CLI Reference](../usage/cli/client/agent/agent.mdx): agent fleet
  commands
- [Node CLI Reference](../usage/cli/client/node/node.mdx): node job commands
- [System Facts](system-facts.md): fact collection and `@fact.*` references
- [API Reference](/gen/api/status-api-status-operations): REST API documentation
- [Job System](job-system.md): how async job processing works
- [Architecture](../architecture/architecture.md): system design overview
