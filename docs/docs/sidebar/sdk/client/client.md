---
sidebar_position: 1
---

# Client Library

The `osapi` package provides a typed Go client for the OSAPI REST API. Create a
client with `New()` and use domain-specific services to interact with the API.

## Quick Start

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", "your-jwt-token")

resp, err := c.Hostname.Get(ctx, "_any")
```

## How the services are made

`internal/sdk/gen` writes the services and their types from the spec of the API,
in `go generate`. Each group of the API is a service, and each operation is a
method:

- The service has the words of the group's path, such as `NetworkDNS` for
  `/api/node/{hostname}/network/dns`.
- The method has the name of the method in the group's definition, such as
  `Get`. The CLI command has the same name.
- A type of an answer has the name of its schema, such as `NetworkDnsEntry`. A
  `{job_id, results}` answer is a `Collection` of its item.
- A method with a body takes the target, then one input struct, such as
  `NetworkDNSUpdateInput`: the keys of the path and the fields of the body.
- A method with no body takes the parameters of the path, then a struct of the
  parameters of the query that a request can leave out.
- Each answer outside 2xx is one of the error types of the SDK.

The methods that upload a file, `File.Upload` and `File.Update`, and
`File.GetChanged` are written by hand.

## Services

| Service                          | Description                  |
| -------------------------------- | ---------------------------- |
| [Service](services/service.md)   | Service management (systemd) |
| [Schedule](services/schedule.md) | Scheduled entry management   |

### Software

| Service                                        | Description        |
| ---------------------------------------------- | ------------------ |
| [Package, UpdatePackages](software/package.md) | Package management |

### Config

| Service                        | Description                 |
| ------------------------------ | --------------------------- |
| [Hostname](config/hostname.md) | Hostname query and update   |
| [Sysctl](config/sysctl.md)     | Kernel parameter management |
| [NTP](config/ntp.md)           | NTP server configuration    |
| [Timezone](config/timezone.md) | System timezone             |

### Node

| Service                            | Description                            |
| ---------------------------------- | -------------------------------------- |
| [Power](node/power.md)             | Power management (reboot, shutdown)    |
| [Process](node/process.md)         | Process management (list, get, signal) |
| [LogEntry, LogSource](node/log.md) | Log query (journal entries, by unit)   |
| [Status](node/status.md)           | Full node status                       |
| [Load](node/load.md)               | Load averages                          |
| [Uptime](node/uptime.md)           | System uptime                          |
| [OS](node/os.md)                   | Operating system info                  |

### Networking

| Service                                           | Description                        |
| ------------------------------------------------- | ---------------------------------- |
| [NetworkDNS](networking/dns.md)                   | DNS configuration query and update |
| [NetworkConnectivity](networking/connectivity.md) | Ping an address from a host        |
| [NetworkInterface](networking/interface.md)       | Network interface configuration    |
| [NetworkInterfaceRoute](networking/route.md)      | Static route configuration         |

### Security

| Service                                | Description               |
| -------------------------------------- | ------------------------- |
| [User, UserSSHKey](security/user.md)   | User account management   |
| [Group](security/group.md)             | Group management          |
| [Certificate](security/certificate.md) | CA certificate management |

### Containers

| Service                                                 | Description         |
| ------------------------------------------------------- | ------------------- |
| [ContainerDocker, ContainerImage](containers/docker.md) | Container lifecycle |

### Files

| Service                          | Description                    |
| -------------------------------- | ------------------------------ |
| [File](files/file.md)            | File management (Object Store) |
| [NodeFile](files/file_deploy.md) | File deployment to agents      |

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

| Service                        | Description                            |
| ------------------------------ | -------------------------------------- |
| [Agent, Facts](agent/agent.md) | Agent discovery, lifecycle, enrollment |

### Health

| Service                                      | Description             |
| -------------------------------------------- | ----------------------- |
| [Health, Controller, NATS](health/health.md) | Health check operations |

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
