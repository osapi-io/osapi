---
sidebar_position: 27
---

# Network interface management

osapi manages the network interface configuration and the static routes of
managed hosts. It configures interfaces and routes with Netplan drop-in files.
Thus the configuration stays after a reboot, and it works with the network
management of Ubuntu.

## How it works

The interface and route providers write Netplan drop-in configuration files to
`/etc/netplan/` on Debian-family systems. Each managed resource uses an `osapi-`
filename prefix to identify OSAPI-owned files:

- **Interfaces**: `osapi-{name}.yaml` configures the addresses, DHCP, gateway,
  MTU, MAC and Wake-on-LAN of the named interface.
- **Routes**: `osapi-{name}-routes.yaml` configures the static routes of the
  named interface.

After it writes each file, the agent runs `netplan generate` to validate the
syntax before it applies the file. If validation fails, the agent puts back the
previous file, and the job returns an error. Thus a bad configuration does not
stop the network connection.

### DNS delete

The DNS provider also supports delete: removes the OSAPI-managed
`/etc/netplan/osapi-dns.yaml` for the interface and runs `netplan apply`.
Returns `changed: true` if the file existed.

## Safety features

- **Validation with netplan generate.** All writes run `netplan generate` before
  they apply. If the generated configuration is not valid, the agent puts back
  the previous file, and the job returns an error. The previous configuration
  stays.
- **Protection of the default route.** The route provider does not delete a
  route if that removes the default gateway. Thus the host does not lose its
  connection.
- **The osapi prefix.** osapi manages only files with the `osapi-` prefix. It
  does not touch system files from installers or other tools.
- **SHA idempotency.** `update` calculates the SHA of the new content and
  compares it with the SHA of the file on disk. Only when they are the same does
  it skip the write and return `changed: false`. A configuration that somebody
  changed by hand is different from the content that osapi writes. Thus osapi
  puts it back, and does not report it as unchanged.
- **Atomic writes.** The agent writes the configuration to a temporary file next
  to the target, and renames it over the target. Thus `netplan generate` never
  reads a file that is only partly written.

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

## CLI usage

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
  --target web-01 --interface eth0

# Broadcast interface create to all hosts
osapi client node network interface create \
  --target _all --name eth1 --dhcp4
```

All commands accept `--json` to give the raw JSON output.

## Broadcast support

All interface and route operations accept broadcast targets. Use `--target _all`
to manage the network configuration on all registered agents. Use a label
selector, for example `--target group:web`, to target a part of the fleet.

Responses always include per-host results:

```
  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME    CHANGED
  web-01    eth0    true
  web-02    eth0    true
```

Skipped and failed hosts appear with their error in the output.

## Supported platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On a platform that is not supported, interface and route operations return
`status: skipped`. They do not fail. Refer to
[Platform Detection](../sdk/platform/detection.md) for how osapi finds the OS
family.

## Container behavior

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

- [CLI Reference](../usage/cli/client/node/network/interface/interface.mdx):
  interface commands
- [CLI Reference](../usage/cli/client/node/network/route/route.mdx): route
  commands
- [SDK Reference](../sdk/client/networking/interface.md): Interface service
- [SDK Reference](../sdk/client/networking/route.md): Route service
- [Network Management](network-management.md): DNS and ping
- [Platform Detection](../sdk/platform/detection.md): OS family detection
- [Configuration](../usage/configuration.md): full configuration reference
