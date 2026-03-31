---
sidebar_position: 1
---

# Client Library

The `osapi` package provides a typed Go client for the OSAPI REST API. Create a
client with `New()` and use domain-specific services to interact with the API.

## Quick Start

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

client := client.New("http://localhost:8080", "your-jwt-token")

resp, err := client.Hostname.Get(ctx, "_any")
```

## Services

### Node Info

| Service                           | Description               |
| --------------------------------- | ------------------------- |
| [Status](node-info/status.md)     | Full node status          |
| [Hostname](node-info/hostname.md) | Hostname query and update |
| [Disk](node-info/disk.md)         | Disk usage                |
| [Memory](node-info/memory.md)     | Memory statistics         |
| [Load](node-info/load.md)         | Load averages             |
| [Uptime](node-info/uptime.md)     | System uptime             |
| [OS](node-info/os.md)             | Operating system info     |

### Network

| Service                 | Description                        |
| ----------------------- | ---------------------------------- |
| [DNS](network/dns.md)   | DNS configuration query and update |
| [Ping](network/ping.md) | Network ping                       |

### System Config

| Service                               | Description                 |
| ------------------------------------- | --------------------------- |
| [Sysctl](system-config/sysctl.md)     | Kernel parameter management |
| [NTP](system-config/ntp.md)           | NTP server configuration    |
| [Timezone](system-config/timezone.md) | System timezone             |

### Operations

| Service                          | Description                            |
| -------------------------------- | -------------------------------------- |
| [Command](operations/command.md) | Command execution (exec, shell)        |
| [Power](operations/power.md)     | Power management (reboot, shutdown)    |
| [Process](operations/process.md) | Process management (list, get, signal) |

### Containers & Scheduling

| Service                                   | Description              |
| ----------------------------------------- | ------------------------ |
| [Docker](containers-scheduling/docker.md) | Container lifecycle      |
| [Cron](containers-scheduling/cron.md)     | Cron schedule management |

### Files

| Service                            | Description                    |
| ---------------------------------- | ------------------------------ |
| [File](files/file.md)              | File management (Object Store) |
| [FileDeploy](files/file_deploy.md) | File deployment to agents      |

### Management

| Service                        | Description                    |
| ------------------------------ | ------------------------------ |
| [Agent](management/agent.md)   | Agent discovery, drain/undrain |
| [Job](management/job.md)       | Async job queue operations     |
| [Health](management/health.md) | Health check operations        |
| [Audit](management/audit.md)   | Audit log operations           |
| [User](management/user.md)     | User account management        |
| [Group](management/group.md)   | Group management               |

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
