---
sidebar_position: 5
---

# System facts

Agents collect **system facts** automatically. A fact is a typed property of the
host that the agent runs on. The agent publishes the facts to a NATS KV bucket.
Facts are used in two features: the `agent get` output, and **fact references**
(`@fact.*`), which put live system values in job parameters.

## What the agent collects

The agent collects the facts with [gohai](https://github.com/osapi-io/gohai)
every 60 seconds: its kernel, hostname, cpu, init, package_mgr, network and
virtualization collectors. `agent.facts.interval` sets this. The agent stores
the facts in the `agent-facts` KV bucket, with a 5-minute TTL.

| Fact              | Description                                        | Example value        |
| ----------------- | -------------------------------------------------- | -------------------- |
| Architecture      | CPU architecture, as `uname -m` prints it          | `x86_64`, `arm64`    |
| Kernel Version    | OS kernel version string                           | `6.8.0-51-generic`   |
| FQDN              | Fully qualified domain name                        | `web-01.example.com` |
| CPU Count         | Number of logical CPUs                             | `8`                  |
| Service Manager   | Init system                                        | `systemd`, `launchd` |
| Package Manager   | System package manager                             | `apt`, `brew`        |
| Interfaces        | Network interfaces with IPv4, IPv6, MAC and family | Refer to below       |
| Primary Interface | Interface name of the default route                | `eth0`, `en0`        |
| Routes            | IP routing table entries                           | Refer to below       |

### Network interfaces

Each interface entry has:

- **Name**: the interface name, for example `eth0` or `en0`
- **IPv4**: the IPv4 address, if it has one
- **IPv6**: the IPv6 address, if it has one
- **MAC**: the hardware address
- **Family**: `inet`, `inet6` or `dual`

The list has only interfaces that are up and that are not loopback interfaces.

### Routes

Each route entry has:

- **Destination**: the target network. The default route is `0.0.0.0`
- **Gateway**: the next-hop address
- **Interface**: the name of the outgoing interface
- **Mask**: the CIDR mask, for example `/24`. The default route is `/0`
- **Metric**: the route metric (Linux only)

The list has only IPv4 routes, on Linux and on macOS.

## Fact references (`@fact.*`)

Fact references put live system values in job parameters. When an agent
processes a job, it replaces each `@fact.*` token in the request data with the
value from its cached facts. The CLI and the API send the `@fact.*` string as it
is. The agent resolves it before it runs the operation.

### Available references

| Reference                 | Resolves to                 | Example value        |
| ------------------------- | --------------------------- | -------------------- |
| `@fact.hostname`          | The hostname of the agent   | `web-01`             |
| `@fact.arch`              | CPU architecture            | `x86_64`             |
| `@fact.kernel`            | Kernel version              | `6.8.0-51-generic`   |
| `@fact.fqdn`              | Fully qualified hostname    | `web-01.example.com` |
| `@fact.interface.primary` | The default route interface | `eth0`               |
| `@fact.custom.<key>`      | A custom fact value         | _(user-defined)_     |

### Usage examples

Get the DNS configuration of the primary network interface:

```bash
osapi client node network dns get \
  --interface-name @fact.interface.primary
```

Print the hostname on the remote host:

```bash
osapi client node command exec \
  --command echo --args "@fact.hostname"
```

Use more than one reference in one command:

```bash
osapi client node command exec \
  --command echo \
  --args "@fact.interface.primary on @fact.hostname"
```

Use fact references with a broadcast target:

```bash
osapi client node command exec \
  --command ip --args "addr,show,dev,@fact.interface.primary" \
  --target _all
```

### How it works

1. The CLI sends the `@fact.*` string in the job request data.
1. The API server publishes the job to NATS without a change.
1. The agent receives the job and looks for `@fact.*` tokens in the request
   data.
1. The agent resolves each token from its local cached facts.
1. The agent gives the resolved data to the provider, which runs the operation.

The agent resolves the references. Thus fact references work with broadcast
(`_all`) and with label targets: each agent puts in its own values. For example,
`@fact.interface.primary` is `eth0` on one host and `en0` on a different host.

If a fact is not available, the job fails with an error that names the reference
that the agent could not resolve. For example, the agent did not collect facts
yet, or the fact key does not exist.

### Supported contexts

Fact references work in **all string values** of an API request that goes to an
agent. The agent resolves all `@fact.*` tokens before it gives the data to the
provider. No endpoint needs special code. If a field accepts a string, it
accepts a fact reference.

This includes:

- **Single strings**: `--interface @fact.interface.primary`
- **Arrays**: `--servers "1.1.1.1,@fact.custom.dns_server"` resolves the fact
  and keeps the IP address as it is
- **Nested maps**: the agent resolves references in JSON objects at all levels
- **All domains**: commands, DNS, sysctl, services, packages and the other
  domains

```bash
# Fact reference in an array field
osapi client node network dns update \
  --servers 1.1.1.1 --servers @fact.custom.backup_dns \
  --interface-name @fact.interface.primary

# Each agent resolves its own facts, so broadcast works:
# web-01 might resolve @fact.interface.primary to "eth0"
# web-02 might resolve it to "ens3"
```

The agent does not change values that are not strings, for example numbers and
booleans.

## See the facts

Use `agent get` to see all the facts of one agent:

```bash
osapi client agent get --hostname web-01
```

The output has the architecture, kernel, FQDN, CPUs, service manager, package
manager, network interfaces and routes. Use `--json` to get all the structured
data.

## Configuration

| Key                    | Description                          | Default       |
| ---------------------- | ------------------------------------ | ------------- |
| `agent.facts.interval` | How often the agent collects facts   | `60s`         |
| `nats.facts.bucket`    | Name of the KV bucket for the facts  | `agent-facts` |
| `nats.facts.ttl`       | TTL of the fact entries              | `5m`          |
| `nats.facts.storage`   | Storage backend (`file` or `memory`) | `file`        |

Refer to [Configuration](../usage/configuration.md) for all the settings.

## Platform support

gohai collects the facts, with one call to these collectors, on each Linux
family and on macOS: Fedora, RHEL and Arch get the same facts as Debian. A
feature that is not supported on a family still reports `skipped`.

| Fact              | gohai collector  | Linux reads                         | macOS reads             |
| ----------------- | ---------------- | ----------------------------------- | ----------------------- |
| Architecture      | `kernel`         | `uname -m`                          | `uname -m`              |
| Kernel Version    | `kernel`         | `uname -r`                          | `uname -r`              |
| FQDN              | `hostname`       | `hostname -s`, then DNS             | `hostname -s`, then DNS |
| CPU Count         | `cpu`            | `/proc/cpuinfo`                     | `sysctl hw.logicalcpu`  |
| Service Manager   | `init`           | the init process                    | `launchd`               |
| Package Manager   | `package_mgr`    | the manager of the family on `PATH` | `brew`, `port`          |
| Interfaces        | `network`        | `/sys/class/net`, `ip addr`         | `ifconfig`              |
| Primary Interface | `network`        | `ip route`                          | `netstat -rn`           |
| Routes            | `network`        | `ip route`                          | `netstat -rn`           |
| Containerized     | `virtualization` | Ohai's container cascade            | Ohai's cascade          |

A collector error does not stop the agent. If a collector fails, the agent still
publishes the facts that it collected. Thus `@fact.interface.primary` can be
unavailable when route collection fails, but `@fact.hostname` and `@fact.arch`
still work.

## Related

- [Agent CLI Reference](../usage/cli/client/agent/agent.mdx): see the facts of
  an agent
- [Command Execution](command-execution.md): use `@fact.*` in commands
- [Network Management](network-management.md): use `@fact.*` in DNS queries
- [Node Management](node-management.md): an overview of agent and node commands
- [Configuration](../usage/configuration.md): the facts interval and the KV
  settings
