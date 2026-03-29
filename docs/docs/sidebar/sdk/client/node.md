---
sidebar_position: 8
---

# NodeService

Node management, network configuration, and command execution. This is the
largest service -- it combines node info, network, and command operations that
all target a specific host.

## Methods

### Node Info

| Method                              | Description                               |
| ----------------------------------- | ----------------------------------------- |
| `Status(ctx, target)`               | Full node status (OS, disk, memory, load) |
| `Hostname(ctx, target)`             | Get system hostname                       |
| `UpdateHostname(ctx, target, name)` | Set system hostname                       |
| `Disk(ctx, target)`                 | Get disk usage                            |
| `Memory(ctx, target)`               | Get memory statistics                     |
| `Load(ctx, target)`                 | Get load averages                         |
| `OS(ctx, target)`                   | Get operating system info                 |
| `Uptime(ctx, target)`               | Get uptime                                |

### Network

| Method                                           | Description        |
| ------------------------------------------------ | ------------------ |
| `GetDNS(ctx, target, iface)`                     | Get DNS config     |
| `UpdateDNS(ctx, target, iface, servers, search)` | Update DNS servers |
| `Ping(ctx, target, address)`                     | Ping a host        |

### Command

| Method            | Description                                 |
| ----------------- | ------------------------------------------- |
| `Exec(ctx, req)`  | Execute a command directly (no shell)       |
| `Shell(ctx, req)` | Execute via `/bin/sh -c` (pipes, redirects) |

### File

| Method                          | Description                         |
| ------------------------------- | ----------------------------------- |
| `FileDeploy(ctx, opts)`         | Deploy file to agent with SHA check |
| `FileStatus(ctx, target, path)` | Check deployed file status          |

See [`FileService`](file.md) for Object Store operations (upload, list, get,
delete) and `FileDeployOpts` details.

## Usage

```go
// Get hostname
resp, err := client.Node.Hostname(ctx, "_any")

// Update hostname
resp, err := client.Node.UpdateHostname(ctx, "web-01", "new-hostname")

// Get disk usage from all hosts
resp, err := client.Node.Disk(ctx, "_all")

// Update DNS
resp, err := client.Node.UpdateDNS(
    ctx, "web-01", "eth0",
    []string{"8.8.8.8", "8.8.4.4"},
    nil,
)

// Execute a command
resp, err := client.Node.Exec(ctx, client.ExecRequest{
    Command: "apt",
    Args:    []string{"install", "-y", "nginx"},
    Target:  "_all",
})

// Execute a shell command
resp, err := client.Node.Shell(ctx, client.ShellRequest{
    Command: "ps aux | grep nginx",
    Target:  "_any",
})

// Deploy a file
resp, err := client.Node.FileDeploy(ctx, client.FileDeployOpts{
    ObjectName:  "nginx.conf",
    Path:        "/etc/nginx/nginx.conf",
    ContentType: "raw",
    Mode:        "0644",
    Target:      "web-01",
})

// Check file status
resp, err := client.Node.FileStatus(
    ctx, "web-01", "/etc/nginx/nginx.conf",
)
```

## Examples

See
[`examples/sdk/client/node.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/node.go)
for node info, and
[`examples/sdk/client/network.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/network.go)
and
[`examples/sdk/client/command.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/command.go)
for network and command examples.

## Result Status

Every result type returned by node operations includes a `Status` field with one
of three values:

| Value     | Meaning                                             |
| --------- | --------------------------------------------------- |
| `ok`      | Operation completed successfully                    |
| `failed`  | Operation failed with an error                      |
| `skipped` | Operation not supported on this OS family or target |

In broadcast responses, any host that does not support the operation (e.g., a
Darwin host in a Linux fleet) appears as `skipped` with an error description.

## Permissions

Node info requires `node:read`. Hostname update requires `node:write`. Network
read requires `network:read`. DNS updates require `network:write`. Commands
require `command:execute`. File deploy requires `file:write`. File status
requires `file:read`.
