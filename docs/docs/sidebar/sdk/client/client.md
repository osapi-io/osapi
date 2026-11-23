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

<!-- osapi:sdk-services -->

### Node

| Service                                                  | Description                                               |
| -------------------------------------------------------- | --------------------------------------------------------- |
| [Certificate](node/certificate.md)                       | The certificate authorities that the host trusts.         |
| [Command](node/command.md)                               | Run a command on target hosts.                            |
| [ContainerDocker](node/container_docker.md)              | The Docker containers of the host.                        |
| [ContainerImage](node/container_image.md)                | The Docker images of the host.                            |
| [Disk](node/disk.md)                                     | Get the disk usage of target hosts.                       |
| [Group](node/group.md)                                   | The groups of the host.                                   |
| [Hostname](node/hostname.md)                             | Get or set the hostname of target hosts.                  |
| [Load](node/load.md)                                     | Get the load averages of target hosts.                    |
| [LogEntry](node/log_entry.md)                            | The entries of the systemd journal.                       |
| [LogSource](node/log_source.md)                          | The sources that write to the systemd journal.            |
| [Memory](node/memory.md)                                 | Get the memory usage of target hosts.                     |
| [NTP](node/ntp.md)                                       | The NTP servers of chrony, and its sync status.           |
| [NetworkConnectivity](node/network_connectivity.md)      | Whether the host reaches an address, by ICMP ping.        |
| [NetworkDNS](node/network_dns.md)                        | The DNS resolvers of each interface of the host.          |
| [NetworkInterface](node/network_interface.md)            | The network interfaces of the host.                       |
| [NetworkInterfaceRoute](node/network_interface_route.md) | The routes of an interface.                               |
| [NodeFile](node/node_file.md)                            | Files of the object store, deployed to the host.          |
| [OS](node/os.md)                                         | Get the operating system of target hosts.                 |
| [Package](node/package.md)                               | The packages installed on the host.                       |
| [Power](node/power.md)                                   | Reboot or shut down the host.                             |
| [Process](node/process.md)                               | The processes of the host.                                |
| [Schedule](node/schedule.md)                             | The cron entries that osapi deploys.                      |
| [Service](node/service.md)                               | The systemd services of the host.                         |
| [Status](node/status.md)                                 | The status summary of a node.                             |
| [Sysctl](node/sysctl.md)                                 | Kernel parameters, written to sysctl.d files and applied. |
| [Timezone](node/timezone.md)                             | The timezone of the host.                                 |
| [UpdatePackages](node/update_packages.md)                | The updates the package index has for the host.           |
| [Uptime](node/uptime.md)                                 | Get the uptime of target hosts.                           |
| [User](node/user.md)                                     | The users of the host.                                    |
| [UserSSHKey](node/user_ssh_key.md)                       | The authorized SSH keys of a user.                        |

### Controller

| Service                                | Description                                                          |
| -------------------------------------- | -------------------------------------------------------------------- |
| [Admission](controller/admission.md)   | Admit, reject and revoke the nodes that ask to join.                 |
| [Agent](controller/agent.md)           | Operations related to agent discovery and details.                   |
| [Audit](controller/audit.md)           | Audit log endpoints for viewing API activity.                        |
| [Controller](controller/controller.md) | The controller of the fleet.                                         |
| [Facts](controller/facts.md)           | Fact key discovery for @fact. references.                            |
| [File](controller/file.md)             | Object Store file management operations.                             |
| [Health](controller/health.md)         | Health check endpoints for liveness, readiness, and detailed status. |
| [Job](controller/job.md)               | Operations related to the job queue.                                 |
| [NATS](controller/nats.md)             | The NATS server of the fleet.                                        |
| [Telemetry](controller/telemetry.md)   | Telemetry endpoints, such as the spans of a trace.                   |

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
