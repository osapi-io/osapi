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

| Service                      | Description                             |
| ---------------------------- | --------------------------------------- |
| [Agent](management/agent.md)                    | Agent discovery, details, drain/undrain |
| [Audit](management/audit.md)                    | Audit log operations                    |
| [Command](operations/command.md)                | Command execution (exec, shell)         |
| [Cron](containers-scheduling/cron.md)           | Cron schedule management                |
| [Disk](node-info/disk.md)                       | Disk usage queries                      |
| [DNS](network/dns.md)                           | DNS configuration query and update      |
| [Docker](containers-scheduling/docker.md)       | Container lifecycle management          |
| [File](files/file.md)                           | File management (Object Store)          |
| [FileDeploy](files/file_deploy.md)              | File deployment to agents               |
| [Health](management/health.md)                  | Health check operations                 |
| [Hostname](node-info/hostname.md)               | Hostname query and update               |
| [Job](management/job.md)                        | Async job queue operations              |
| [Load](node-info/load.md)                       | Load average queries                    |
| [Memory](node-info/memory.md)                   | Memory statistics queries               |
| [NTP](system-config/ntp.md)                     | NTP configuration management            |
| [OS](node-info/os.md)                           | Operating system info queries           |
| [Ping](network/ping.md)                         | Network ping operations                 |
| [Power](operations/power.md)                    | Power management (reboot, shutdown)     |
| [Status](node-info/status.md)                   | Aggregated node status                  |
| [Sysctl](system-config/sysctl.md)               | Kernel parameter management             |
| [Timezone](system-config/timezone.md)           | System timezone management              |
| [Uptime](node-info/uptime.md)                   | Uptime queries                          |

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
