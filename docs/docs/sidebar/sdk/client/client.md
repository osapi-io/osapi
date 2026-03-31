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

| Service                        | Description                                  |
| ------------------------------ | -------------------------------------------- |
| [Agent](agent.md)              | Agent discovery, details, drain/undrain       |
| [Audit](audit.md)              | Audit log operations                         |
| [Command](command.md)          | Command execution (exec, shell)              |
| [Cron](cron.md)                | Cron schedule management                     |
| [Disk](disk.md)                | Disk usage queries                           |
| [DNS](dns.md)                  | DNS configuration query and update           |
| [Docker](docker.md)            | Container lifecycle management               |
| [File](file.md)                | File management (Object Store)               |
| [FileDeploy](file_deploy.md)   | File deployment to agents                    |
| [Health](health.md)            | Health check operations                      |
| [Hostname](hostname.md)        | Hostname query and update                    |
| [Job](job.md)                  | Async job queue operations                   |
| [Load](load.md)                | Load average queries                         |
| [Memory](memory.md)            | Memory statistics queries                    |
| [NTP](ntp.md)                  | NTP configuration management                |
| [OS](os.md)                    | Operating system info queries                |
| [Ping](ping.md)                | Network ping operations                      |
| [Status](status.md)            | Aggregated node status                       |
| [Sysctl](sysctl.md)            | Kernel parameter management                  |
| [Timezone](timezone.md)        | System timezone management                   |
| [Uptime](uptime.md)            | Uptime queries                               |

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
