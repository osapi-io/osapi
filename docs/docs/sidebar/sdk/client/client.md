---
sidebar_position: 1
---

# Client Library

The `osapi` package provides a typed Go client for the OSAPI REST API. Create a
client with `New()` and use domain-specific services to interact with the API.

## Quick Start

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

client := client.New("http://localhost:8080", "your-jwt-token")

resp, err := client.Hostname.Get(ctx, "_any")
```

## Services

| Service                        | Description                  |
| ------------------------------ | ---------------------------- |
| [Service](services/service.md) | Service management (systemd) |
| [Cron](services/cron.md)       | Cron schedule management     |

### Software

| Service                        | Description        |
| ------------------------------ | ------------------ |
| [Package](software/package.md) | Package management |

### Config

| Service                        | Description                 |
| ------------------------------ | --------------------------- |
| [Hostname](config/hostname.md) | Hostname query and update   |
| [Sysctl](config/sysctl.md)     | Kernel parameter management |
| [NTP](config/ntp.md)           | NTP server configuration    |
| [Timezone](config/timezone.md) | System timezone             |

### Node

| Service                    | Description                            |
| -------------------------- | -------------------------------------- |
| [Power](node/power.md)     | Power management (reboot, shutdown)    |
| [Process](node/process.md) | Process management (list, get, signal) |
| [Log](node/log.md)         | Log query (journal entries, by unit)   |
| [Status](node/status.md)   | Full node status                       |
| [Load](node/load.md)       | Load averages                          |
| [Uptime](node/uptime.md)   | System uptime                          |
| [OS](node/os.md)           | Operating system info                  |

### Networking

| Service                              | Description                        |
| ------------------------------------ | ---------------------------------- |
| [DNS](networking/dns.md)             | DNS configuration query and update |
| [Ping](networking/ping.md)           | Network ping                       |
| [Interface](networking/interface.md) | Network interface configuration    |
| [Route](networking/route.md)         | Static route configuration         |

### Security

| Service                                | Description               |
| -------------------------------------- | ------------------------- |
| [User](security/user.md)               | User account management   |
| [Group](security/group.md)             | Group management          |
| [Certificate](security/certificate.md) | CA certificate management |

### Containers

| Service                        | Description         |
| ------------------------------ | ------------------- |
| [Docker](containers/docker.md) | Container lifecycle |

### Files

| Service                            | Description                    |
| ---------------------------------- | ------------------------------ |
| [File](files/file.md)              | File management (Object Store) |
| [FileDeploy](files/file_deploy.md) | File deployment to agents      |

### Command

| Service                       | Description                     |
| ----------------------------- | ------------------------------- |
| [Command](command/command.md) | Command execution (exec, shell) |

### Hardware

| Service                      | Description       |
| ---------------------------- | ----------------- |
| [Disk](hardware/disk.md)     | Disk usage        |
| [Memory](hardware/memory.md) | Memory statistics |

### Audit

| Service                 | Description          |
| ----------------------- | -------------------- |
| [Audit](audit/audit.md) | Audit log operations |

### Jobs

| Service            | Description                |
| ------------------ | -------------------------- |
| [Job](jobs/job.md) | Async job queue operations |

### Agent

| Service                 | Description                            |
| ----------------------- | -------------------------------------- |
| [Agent](agent/agent.md) | Agent discovery, lifecycle, enrollment |

### Health

| Service                    | Description             |
| -------------------------- | ----------------------- |
| [Health](health/health.md) | Health check operations |

## Client Options

| Option                         | Description                    |
| ------------------------------ | ------------------------------ |
| `WithLogger(logger)`           | Set custom `slog.Logger`       |
| `WithHTTPTransport(transport)` | Set custom `http.RoundTripper` |

`WithLogger` defaults to `slog.Default()`. `WithHTTPTransport` sets the base
transport for HTTP requests.

## Targeting

Most operations accept a `target` parameter:

| Target      | Behavior                                    |
| ----------- | ------------------------------------------- |
| `_any`      | Send to any available agent (load balanced) |
| `_all`      | Broadcast to every agent                    |
| `hostname`  | Send to a specific host                     |
| `key:value` | Send to agents matching a label             |
