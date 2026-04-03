---
sidebar_position: 5
---

# System Facts

Agents automatically collect **system facts** — typed properties about the host
they run on — and publish them to a dedicated NATS KV bucket. Facts power two
features: the `agent get` display and **fact references** (`@fact.*`) that let
you inject live system values into job parameters.

## What Gets Collected

Facts are gathered from providers every 60 seconds (configurable via
`agent.facts.interval`) and stored in the `agent-facts` KV bucket with a
5-minute TTL.

| Fact              | Description                                     | Example Value        |
| ----------------- | ----------------------------------------------- | -------------------- |
| Architecture      | CPU architecture                                | `amd64`, `arm64`     |
| Kernel Version    | OS kernel version string                        | `6.8.0-51-generic`   |
| FQDN              | Fully qualified domain name                     | `web-01.example.com` |
| CPU Count         | Number of logical CPUs                          | `8`                  |
| Service Manager   | Init system                                     | `systemd`, `launchd` |
| Package Manager   | System package manager                          | `apt`, `brew`        |
| Interfaces        | Network interfaces with IPv4, IPv6, MAC, family | See below            |
| Primary Interface | Interface name of the default route             | `eth0`, `en0`        |
| Routes            | IP routing table entries                        | See below            |

### Network Interfaces

Each interface entry includes:

- **Name** — interface name (e.g., `eth0`, `en0`)
- **IPv4** — IPv4 address (if assigned)
- **IPv6** — IPv6 address (if assigned)
- **MAC** — hardware address
- **Family** — `inet`, `inet6`, or `dual`

Only non-loopback, up interfaces are included.

### Routes

Each route entry includes:

- **Destination** — target network or `default` / `0.0.0.0`
- **Gateway** — next-hop address
- **Interface** — outgoing interface name
- **Mask** — CIDR mask (Linux only, e.g., `/24`)
- **Metric** — route metric (Linux only)
- **Flags** — route flags

## Fact References (`@fact.*`)

Fact references let you use live system values in job parameters. When an agent
processes a job, it replaces any `@fact.*` token in the request data with the
corresponding value from its cached facts. This happens transparently — the CLI
and API send the literal `@fact.*` string, and the agent resolves it before
executing the operation.

### Available References

| Reference                 | Resolves To              | Example Value        |
| ------------------------- | ------------------------ | -------------------- |
| `@fact.hostname`          | Agent's hostname         | `web-01`             |
| `@fact.arch`              | CPU architecture         | `amd64`              |
| `@fact.kernel`            | Kernel version           | `6.8.0-51-generic`   |
| `@fact.fqdn`              | Fully qualified hostname | `web-01.example.com` |
| `@fact.interface.primary` | Default route interface  | `eth0`               |
| `@fact.custom.<key>`      | Custom fact value        | _(user-defined)_     |

### Usage Examples

Query DNS configuration on the primary network interface:

```bash
osapi client node network dns get \
  --interface-name @fact.interface.primary
```

Echo the hostname on the remote host:

```bash
osapi client node command exec \
  --command echo --args "@fact.hostname"
```

Use multiple references in a single command:

```bash
osapi client node command exec \
  --command echo \
  --args "@fact.interface.primary on @fact.hostname"
```

Use fact references with broadcast targeting:

```bash
osapi client node command exec \
  --command ip --args "addr,show,dev,@fact.interface.primary" \
  --target _all
```

### How It Works

1. The CLI sends the literal `@fact.*` string in the job request data
2. The API server publishes the job to NATS as-is
3. The agent receives the job and checks the request data for `@fact.*` tokens
4. Each token is resolved against the agent's locally cached facts
5. The resolved data is passed to the provider for execution

Because resolution happens agent-side, fact references work correctly with
broadcast (`_all`) and label-based routing — each agent substitutes its own
values. For example, `@fact.interface.primary` resolves to `eth0` on one host
and `en0` on another.

If a referenced fact is not available (e.g., the agent hasn't collected facts
yet, or the fact key doesn't exist), the job fails with an error describing
which reference could not be resolved.

### Supported Contexts

Fact references work in **any string value** in any API request that reaches
an agent. The agent resolves all `@fact.*` tokens before passing data to the
provider — no special handling is needed per endpoint. If a field accepts a
string, it accepts a fact reference.

This includes:

- **Standalone strings** — `--interface-name @fact.interface.primary`
- **Arrays** — `--servers "1.1.1.1,@fact.custom.dns_server"` resolves the
  fact while keeping the literal IP
- **Nested maps** — references inside JSON objects are resolved recursively
- **Any domain** — commands, DNS, sysctl, services, packages, etc.

```bash
# Fact reference in an array field
osapi client node network dns update \
  --servers 1.1.1.1 --servers @fact.custom.backup_dns \
  --interface-name @fact.interface.primary

# Each agent resolves its own facts, so broadcast works:
# web-01 might resolve @fact.interface.primary to "eth0"
# web-02 might resolve it to "ens3"
```

Non-string values (numbers, booleans) are not modified.

## Viewing Facts

Use `agent get` to see the full facts for a specific agent:

```bash
osapi client agent get --hostname web-01
```

The output includes architecture, kernel, FQDN, CPUs, service/package manager,
network interfaces, and routes. Use `--json` for the complete structured data.

## Configuration

| Key                    | Description                          | Default       |
| ---------------------- | ------------------------------------ | ------------- |
| `agent.facts.interval` | How often facts are collected        | `60s`         |
| `nats.facts.bucket`    | KV bucket name for facts storage     | `agent-facts` |
| `nats.facts.ttl`       | TTL for facts entries                | `5m`          |
| `nats.facts.storage`   | Storage backend (`file` or `memory`) | `file`        |

See [Configuration](../usage/configuration.md) for the full reference.

## Platform Support

Facts are collected using platform-specific providers. All facts are available
on both Linux and macOS:

| Fact              | Linux Provider            | macOS Provider            |
| ----------------- | ------------------------- | ------------------------- |
| Architecture      | `gopsutil`                | `gopsutil`                |
| Kernel Version    | `gopsutil`                | `gopsutil`                |
| FQDN              | `gopsutil`                | `gopsutil`                |
| CPU Count         | `gopsutil`                | `gopsutil`                |
| Service Manager   | `gopsutil`                | `gopsutil`                |
| Package Manager   | `gopsutil`                | `gopsutil`                |
| Interfaces        | `net.Interfaces` (stdlib) | `net.Interfaces` (stdlib) |
| Primary Interface | `/proc/net/route` parsing | `netstat -rn` parsing     |
| Routes            | `/proc/net/route` parsing | `netstat -rn` parsing     |

Provider errors are non-fatal — if a provider fails, the agent still publishes
whatever facts it could gather. This means `@fact.interface.primary` may be
unavailable if route collection fails, but `@fact.hostname` and `@fact.arch`
will still work.

## Related

- [Agent CLI Reference](../usage/cli/client/agent/agent.mdx) -- view agent facts
- [Command Execution](command-execution.md) -- use `@fact.*` in commands
- [Network Management](network-management.md) -- use `@fact.*` in DNS queries
- [Node Management](node-management.md) -- agent vs. node overview
- [Configuration](../usage/configuration.md) -- facts interval and KV settings
