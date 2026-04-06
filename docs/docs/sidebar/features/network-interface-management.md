---
sidebar_position: 27
---

# Network Interface Management

OSAPI provides management of network interface configuration and static routes
on managed hosts. Interfaces and routes are configured via Netplan drop-in
files, giving persistent, reboot-safe configuration that integrates naturally
with Ubuntu's network management stack.

## How It Works

The interface and route providers write Netplan drop-in configuration files to
`/etc/netplan/` on Debian-family systems. Each managed resource uses an `osapi-`
filename prefix to identify OSAPI-owned files:

- **Interfaces**: `osapi-{name}.yaml` — configures addresses, DHCP, gateway,
  MTU, MAC, and Wake-on-LAN for the named interface.
- **Routes**: `osapi-{name}-routes.yaml` — configures static routes for the
  named interface.

After writing each file, the agent runs `netplan generate` to validate syntax
before applying. If validation fails, the file is rolled back and the job
returns an error. This prevents bad configuration from breaking network
connectivity.

### Interface Operations

**List** — Returns all interfaces known to Netplan on the host. Each entry
includes the interface name, DHCP settings, IP addresses, gateway, MTU, MAC
address, Wake-on-LAN state, managed flag (whether OSAPI manages the interface),
and state.

**Get** — Returns the configuration of a single interface by name.

**Create** — Writes a new Netplan drop-in file for the interface. Idempotent:
returns `changed: false` if already managed. Use `update` to replace the
configuration. Runs `netplan generate` to validate, then `netplan apply` to
activate. Returns `changed: true` if the file was created.

**Update** — Replaces an existing Netplan drop-in file for the interface. Runs
`netplan generate` and `netplan apply`. Returns `changed: false` if the content
has not changed (same SHA). Fails if no OSAPI-managed configuration exists for
that interface — use `create` first.

**Delete** — Removes the OSAPI-managed Netplan drop-in file for the interface
and runs `netplan apply` to remove the configuration. Returns `changed: true` if
the file existed. Only OSAPI-managed files (with the `osapi-` prefix) can be
deleted.

### Route Operations

**List** — Returns all routes visible in the kernel routing table on the host.
Each entry includes the destination, gateway, interface, metric, and scope.

**Get** — Returns the routes for a specific interface (reads from the
OSAPI-managed Netplan routes file).

**Create** — Writes a new Netplan drop-in routes file for the interface. Each
route specifies a destination in CIDR notation, a gateway IP, and an optional
metric. Idempotent: returns `changed: false` if already managed. Use `update` to
replace it. Runs `netplan generate` to validate.

**Update** — Replaces an existing OSAPI-managed routes file for the interface.
Returns `changed: false` if the content is unchanged.

**Delete** — Removes the OSAPI-managed routes file for the interface and runs
`netplan apply`.

### DNS Delete

The DNS provider also supports delete: removes the OSAPI-managed
`/etc/netplan/osapi-dns.yaml` for the interface and runs `netplan apply`.
Returns `changed: true` if the file existed.

## Safety Features

- **Netplan generate validation** — all writes run `netplan generate` before
  applying. If the generated config is invalid, the file is rolled back and the
  job returns an error. The previous configuration is preserved.
- **Default route protection** — the route provider refuses to delete routes
  that would remove the default gateway, preventing loss of connectivity.
- **OSAPI prefix** — only files with the `osapi-` prefix are managed. System
  files created by installers or other tools are not touched.
- **SHA-based idempotency** — `update` computes a SHA of the new content and
  skips the write if it matches the existing file, returning `changed: false`.

## Operations

| Operation        | Description                                      |
| ---------------- | ------------------------------------------------ |
| Interface List   | List all interfaces on the host                  |
| Interface Get    | Get configuration for a specific interface       |
| Interface Create | Create a new Netplan interface config            |
| Interface Update | Replace an existing Netplan interface config     |
| Interface Delete | Remove an OSAPI-managed interface config         |
| Route List       | List all routes in the kernel routing table      |
| Route Get        | Get managed routes for a specific interface      |
| Route Create     | Create new static routes for an interface        |
| Route Update     | Replace static routes for an interface           |
| Route Delete     | Remove OSAPI-managed routes for an interface     |
| DNS Delete       | Remove OSAPI-managed DNS config for an interface |

## CLI Usage

```bash
# List all interfaces
osapi client node network interface list --target web-01

# Get a specific interface
osapi client node network interface get \
  --target web-01 --name eth0

# Create a static interface config
osapi client node network interface create \
  --target web-01 --name eth0 \
  --address 192.168.1.100/24 \
  --gateway4 192.168.1.1

# Create an interface with DHCP
osapi client node network interface create \
  --target web-01 --name eth0 --dhcp4

# Update an interface (replace config)
osapi client node network interface update \
  --target web-01 --name eth0 \
  --address 192.168.1.200/24 \
  --gateway4 192.168.1.1 \
  --mtu 9000

# Delete an interface config
osapi client node network interface delete \
  --target web-01 --name eth0

# List all routes
osapi client node network route list --target web-01

# Get routes for an interface
osapi client node network route get \
  --target web-01 --interface eth0

# Create static routes
osapi client node network route create \
  --target web-01 --interface eth0 \
  --route 10.0.0.0/8:192.168.1.1 \
  --route 172.16.0.0/12:192.168.1.1:100

# Delete routes for an interface
osapi client node network route delete \
  --target web-01 --interface eth0

# Delete DNS config for an interface
osapi client node network dns delete \
  --target web-01 --interface-name eth0

# Broadcast interface create to all hosts
osapi client node network interface create \
  --target _all --name eth1 --dhcp4
```

All commands support `--json` for raw JSON output.

## Broadcast Support

All interface and route operations support broadcast targeting. Use
`--target _all` to manage network configuration on every registered agent, or
use a label selector like `--target group:web` to target a subset.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME    CHANGED
  web-01    eth0    true
  web-02    eth0    true
```

Skipped and failed hosts appear with their error in the output.

## Supported Platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On unsupported platforms, interface and route operations return
`status: skipped` instead of failing. See
[Platform Detection](../sdk/platform/detection.md) for details on OS family
detection.

## Container Behavior

Interface and route operations return `status: skipped` inside containers.
Netplan requires direct host access and is not available in standard container
environments.

## Permissions

| Operation                        | Permission      |
| -------------------------------- | --------------- |
| Interface List, Get              | `network:read`  |
| Interface Create, Update, Delete | `network:write` |
| Route List, Get                  | `network:read`  |
| Route Create, Update, Delete     | `network:write` |
| DNS Get                          | `network:read`  |
| DNS Update, Delete               | `network:write` |

Network read operations require `network:read`, included in all built-in roles
(`admin`, `write`, `read`). Mutation operations require `network:write`,
included in the `admin` and `write` roles.

## Related

- [CLI Reference](../usage/cli/client/node/network/interface/interface.md) --
  interface commands
- [CLI Reference](../usage/cli/client/node/network/route/route.md) -- route
  commands
- [SDK Reference](../sdk/client/network/interface.md) -- Interface service
- [SDK Reference](../sdk/client/network/route.md) -- Route service
- [Network Management](network-management.md) -- DNS and ping
- [Platform Detection](../sdk/platform/detection.md) -- OS family detection
- [Configuration](../usage/configuration.md) -- full configuration reference
