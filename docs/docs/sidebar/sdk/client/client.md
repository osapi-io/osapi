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

The same run writes a program for each service,
`examples/sdk/client/<service>.go`. It calls each method with the input of each
example of the method's definition, and prints the answer. `just examples-check`
compiles each program, so an example cannot go stale.

## Services

Each service has a page, which `internal/sdk/gen` writes from the same model:
its methods with their signatures, doc comments and permissions, the code of
each example of the method's definition, and its types. The services of a node
act on the hosts of a target. The services of the controller act on the
controller itself.

The pages are in a folder for each area, such as Networking or Fleet. The UI's
Behaviors page has the same areas, in the same order. Each behavior and each
endpoint group declares its area in Go, and `internal/area` gives the label and
the order of each area. The spec carries them, and the generator reads them from
the spec.

<!-- osapi:sdk-services -->

### Services

| Service                        | Description                       |
| ------------------------------ | --------------------------------- |
| [Service](services/service.md) | The systemd services of the host. |

### Software

| Service                                       | Description                                     |
| --------------------------------------------- | ----------------------------------------------- |
| [Package](software/package.md)                | The packages installed on the host.             |
| [UpdatePackages](software/update_packages.md) | The updates the package index has for the host. |

### Config

| Service                         | Description                                               |
| ------------------------------- | --------------------------------------------------------- |
| [Hostname](config/hostname.md)  | Get or set the hostname of target hosts.                  |
| [NTP](config/ntp.md)            | The NTP servers of chrony, and its sync status.           |
| [NodeFile](config/node_file.md) | Files of the object store, deployed to the host.          |
| [Schedule](config/schedule.md)  | The cron entries that osapi deploys.                      |
| [Sysctl](config/sysctl.md)      | Kernel parameters, written to sysctl.d files and applied. |
| [Timezone](config/timezone.md)  | The timezone of the host.                                 |

### Node

| Service                    | Description                               |
| -------------------------- | ----------------------------------------- |
| [Load](node/load.md)       | Get the load averages of target hosts.    |
| [OS](node/os.md)           | Get the operating system of target hosts. |
| [Power](node/power.md)     | Reboot or shut down the host.             |
| [Process](node/process.md) | The processes of the host.                |
| [Status](node/status.md)   | The status summary of a node.             |
| [Uptime](node/uptime.md)   | Get the uptime of target hosts.           |

### Logs

| Service                         | Description                                    |
| ------------------------------- | ---------------------------------------------- |
| [LogEntry](logs/log_entry.md)   | The entries of the systemd journal.            |
| [LogSource](logs/log_source.md) | The sources that write to the systemd journal. |

### Networking

| Service                                                        | Description                                        |
| -------------------------------------------------------------- | -------------------------------------------------- |
| [NetworkConnectivity](networking/network_connectivity.md)      | Whether the host reaches an address, by ICMP ping. |
| [NetworkDNS](networking/network_dns.md)                        | The DNS resolvers of each interface of the host.   |
| [NetworkInterface](networking/network_interface.md)            | The network interfaces of the host.                |
| [NetworkInterfaceRoute](networking/network_interface_route.md) | The routes of an interface.                        |

### Security

| Service                                | Description                                       |
| -------------------------------------- | ------------------------------------------------- |
| [Certificate](security/certificate.md) | The certificate authorities that the host trusts. |
| [Group](security/group.md)             | The groups of the host.                           |
| [User](security/user.md)               | The users of the host.                            |
| [UserSSHKey](security/user_ssh_key.md) | The authorized SSH keys of a user.                |

### Containers

| Service                                           | Description                        |
| ------------------------------------------------- | ---------------------------------- |
| [ContainerDocker](containers/container_docker.md) | The Docker containers of the host. |
| [ContainerImage](containers/container_image.md)   | The Docker images of the host.     |

### Object Store

| Service                     | Description                              |
| --------------------------- | ---------------------------------------- |
| [File](objectstore/file.md) | Object Store file management operations. |

### Command

| Service                       | Description                    |
| ----------------------------- | ------------------------------ |
| [Command](command/command.md) | Run a command on target hosts. |

### Hardware

| Service                      | Description                           |
| ---------------------------- | ------------------------------------- |
| [Disk](hardware/disk.md)     | Get the disk usage of target hosts.   |
| [Memory](hardware/memory.md) | Get the memory usage of target hosts. |

### Audit

| Service                 | Description                                   |
| ----------------------- | --------------------------------------------- |
| [Audit](audit/audit.md) | Audit log endpoints for viewing API activity. |

### Fleet

| Service                         | Description                                                          |
| ------------------------------- | -------------------------------------------------------------------- |
| [Admission](fleet/admission.md) | Admit, reject and revoke the nodes that ask to join.                 |
| [Agent](fleet/agent.md)         | Operations related to agent discovery and details.                   |
| [Health](fleet/health.md)       | Health check endpoints for liveness, readiness, and detailed status. |
| [Job](fleet/job.md)             | Operations related to the job queue.                                 |
| [Trace](fleet/trace.md)         | The spans of a trace.                                                |

<!-- /osapi:sdk-services -->

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
