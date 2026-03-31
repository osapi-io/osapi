---
sidebar_position: 8
---

# Node Services

Node-targeted operations are split into per-domain services. Each service
handles a single concern and is accessed directly on the client.

## Services

### StatusService

| Method             | Description                               |
| ------------------ | ----------------------------------------- |
| `Status.Get(ctx, target)` | Full node status (OS, disk, memory, load) |

### HostnameService

| Method                            | Description        |
| --------------------------------- | ------------------ |
| `Hostname.Get(ctx, target)`       | Get system hostname |
| `Hostname.Update(ctx, target, name)` | Set system hostname |

### DiskService

| Method                  | Description    |
| ----------------------- | -------------- |
| `Disk.Get(ctx, target)` | Get disk usage |

### MemoryService

| Method                    | Description          |
| ------------------------- | -------------------- |
| `Memory.Get(ctx, target)` | Get memory statistics |

### LoadService

| Method                  | Description       |
| ----------------------- | ----------------- |
| `Load.Get(ctx, target)` | Get load averages |

### UptimeService

| Method                    | Description |
| ------------------------- | ----------- |
| `Uptime.Get(ctx, target)` | Get uptime  |

### OSService

| Method                | Description             |
| --------------------- | ----------------------- |
| `OS.Get(ctx, target)` | Get operating system info |

### DNSService

| Method                                       | Description        |
| -------------------------------------------- | ------------------ |
| `DNS.Get(ctx, target, iface)`                | Get DNS config     |
| `DNS.Update(ctx, target, iface, servers, search)` | Update DNS servers |

### PingService

| Method                          | Description |
| ------------------------------- | ----------- |
| `Ping.Do(ctx, target, address)` | Ping a host |

### CommandService

| Method                  | Description                                 |
| ----------------------- | ------------------------------------------- |
| `Command.Exec(ctx, req)` | Execute a command directly (no shell)       |
| `Command.Shell(ctx, req)` | Execute via `/bin/sh -c` (pipes, redirects) |

### FileDeployService

| Method                                | Description                         |
| ------------------------------------- | ----------------------------------- |
| `FileDeploy.Deploy(ctx, opts)`        | Deploy file to agent with SHA check |
| `FileDeploy.Undeploy(ctx, opts)`      | Remove a deployed file              |
| `FileDeploy.Status(ctx, target, path)` | Check deployed file status          |

See [`FileService`](file.md) for Object Store operations (upload, list, get,
delete) and `FileDeployOpts` details.

## Usage

```go
// Get hostname
resp, err := client.Hostname.Get(ctx, "_any")

// Update hostname
resp, err := client.Hostname.Update(ctx, "web-01", "new-hostname")

// Get disk usage from all hosts
resp, err := client.Disk.Get(ctx, "_all")

// Update DNS
resp, err := client.DNS.Update(
    ctx, "web-01", "eth0",
    []string{"8.8.8.8", "8.8.4.4"},
    nil,
)

// Execute a command
resp, err := client.Command.Exec(ctx, client.ExecRequest{
    Command: "apt",
    Args:    []string{"install", "-y", "nginx"},
    Target:  "_all",
})

// Execute a shell command
resp, err := client.Command.Shell(ctx, client.ShellRequest{
    Command: "ps aux | grep nginx",
    Target:  "_any",
})

// Deploy a file
resp, err := client.FileDeploy.Deploy(ctx, client.FileDeployOpts{
    ObjectName:  "nginx.conf",
    Path:        "/etc/nginx/nginx.conf",
    ContentType: "raw",
    Mode:        "0644",
    Target:      "web-01",
})

// Check file status
resp, err := client.FileDeploy.Status(
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

Any host that does not support the operation (e.g., a Darwin host in a Linux
fleet) appears as `skipped` with an error description. This applies to both
single-target and broadcast responses -- the response shape is identical.

## Permissions

Node info requires `node:read`. Hostname update requires `node:write`. Network
read requires `network:read`. DNS updates require `network:write`. Commands
require `command:execute`. File deploy requires `file:write`. File status
requires `file:read`.
