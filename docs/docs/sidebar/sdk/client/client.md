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

| Service                                             | Description                                                          |
| --------------------------------------------------- | -------------------------------------------------------------------- |
| [Admission](admission.md)                           | Admit, reject and revoke the nodes that ask to join.                 |
| [Agent](agent.md)                                   | Operations related to agent discovery and details.                   |
| [Audit](audit.md)                                   | Audit log endpoints for viewing API activity.                        |
| [Certificate](certificate.md)                       | The certificate authorities that the host trusts.                    |
| [Command](command.md)                               | Run a command on target hosts.                                       |
| [ContainerDocker](container_docker.md)              | The Docker containers of the host.                                   |
| [ContainerImage](container_image.md)                | The Docker images of the host.                                       |
| [Controller](controller.md)                         | The controller of the fleet.                                         |
| [Disk](disk.md)                                     | Get the disk usage of target hosts.                                  |
| [Facts](facts.md)                                   | Fact key discovery for @fact. references.                            |
| [File](file.md)                                     | Object Store file management operations.                             |
| [Group](group.md)                                   | The groups of the host.                                              |
| [Health](health.md)                                 | Health check endpoints for liveness, readiness, and detailed status. |
| [Hostname](hostname.md)                             | Get or set the hostname of target hosts.                             |
| [Job](job.md)                                       | Operations related to the job queue.                                 |
| [Load](load.md)                                     | Get the load averages of target hosts.                               |
| [LogEntry](log_entry.md)                            | The entries of the systemd journal.                                  |
| [LogSource](log_source.md)                          | The sources that write to the systemd journal.                       |
| [Memory](memory.md)                                 | Get the memory usage of target hosts.                                |
| [NATS](nats.md)                                     | The NATS server of the fleet.                                        |
| [NTP](ntp.md)                                       | The NTP servers of chrony, and its sync status.                      |
| [NetworkConnectivity](network_connectivity.md)      | Whether the host reaches an address, by ICMP ping.                   |
| [NetworkDNS](network_dns.md)                        | The DNS resolvers of each interface of the host.                     |
| [NetworkInterface](network_interface.md)            | The network interfaces of the host.                                  |
| [NetworkInterfaceRoute](network_interface_route.md) | The routes of an interface.                                          |
| [NodeFile](node_file.md)                            | Files of the object store, deployed to the host.                     |
| [OS](os.md)                                         | Get the operating system of target hosts.                            |
| [Package](package.md)                               | The packages installed on the host.                                  |
| [Power](power.md)                                   | Reboot or shut down the host.                                        |
| [Process](process.md)                               | The processes of the host.                                           |
| [Schedule](schedule.md)                             | The cron entries that osapi deploys.                                 |
| [Service](service.md)                               | The systemd services of the host.                                    |
| [Status](status.md)                                 | The status summary of a node.                                        |
| [Sysctl](sysctl.md)                                 | Kernel parameters, written to sysctl.d files and applied.            |
| [Telemetry](telemetry.md)                           | Telemetry endpoints, such as the spans of a trace.                   |
| [Timezone](timezone.md)                             | The timezone of the host.                                            |
| [UpdatePackages](update_packages.md)                | The updates the package index has for the host.                      |
| [Uptime](uptime.md)                                 | Get the uptime of target hosts.                                      |
| [User](user.md)                                     | The users of the host.                                               |
| [UserSSHKey](user_ssh_key.md)                       | The authorized SSH keys of a user.                                   |

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
