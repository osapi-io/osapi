---
sidebar_position: 27
---

# Network interface management

osapi manages the network interface configuration and the static routes of
managed hosts. It configures interfaces and routes with Netplan drop-in files.
Thus the configuration stays after a reboot, and it works with the network
management of Ubuntu.

## How it works

Interfaces are the
[interface behavior](../reference/behaviors/interface/interface.md). `list` and
`get` read each interface through gohai on every platform: `/sys/class/net` and
`ip` on Linux, and `ifconfig` and `netstat` on macOS. An interface has its
state, MAC address, MTU and addresses. `dhcp4` is true when a route of the
interface has proto `dhcp`. `primary` is true when the interface has the default
route. `list` leaves out the loopback.

Routes are the [route behavior](../reference/behaviors/route/route.md), under an
interface: `osapi client node network interface route`. `list` reads the routes
of the main table that go out of the interface, through gohai on every platform:
`ip route` on Linux, and `netstat` on macOS. A route has its destination,
gateway, interface, family, metric, `proto` (where it came from, such as `dhcp`,
`kernel` or `static`) and scope.

`create`, `update` and `delete` write a Netplan file of the interface's own on
the Debian family, through the internal
[netplan behavior](../reference/behaviors/netplan/netplan.md):

- **Interfaces**: `osapi-{name}.yaml` has the addresses, DHCP, default gateways,
  MTU, MAC and Wake-on-LAN of the named interface. `dhcp4` and `dhcp6` are
  `false` unless you give them, so the interface has only the addresses that you
  give. A gateway is a default route: Netplan deprecates `gateway4` and
  `gateway6`.
- **Routes**: `osapi-{name}-routes.yaml` has the static routes of the named
  interface. Each call is one route, as `ip route add` and `ip route del` are:
  `--destination` in CIDR notation, `--via` an address of the same family, and
  an optional `--metric` from 0 to 4294967295. `create` of a route that osapi
  has, the same, changes nothing; `update` changes the gateway or the metric of
  one; `delete` drops one, and the last one removes the file. osapi refuses a
  default route: set the gateway of the interface with the interface behavior.

osapi writes the config only under a Netplan definition that is the interface's
own. If a match defines the interface, such as the `all` definition of
cloud-init for `en*`, osapi refuses the write: Netplan applies that definition,
so the config would change each interface that the match takes, now and later.
Give the interface a definition of its own first. An interface that the host
does not have is not found: `netplan status` lists only the interfaces that
Netplan manages, so osapi reads the interfaces through gohai.

After it writes each file, the agent runs `netplan generate` to validate the
syntax before it applies the file. If validation fails, the agent puts back the
content that the file had, or removes a new file, and the job returns an error.
Thus a bad configuration does not stop the network connection.

If `netplan apply` fails, the file stays, and the job returns an error. The
record of the file has a mark that tells that the apply did not finish. Thus a
retry with the same input reports a change and runs `netplan apply` again. A
delete whose `netplan apply` fails puts back the file with the mark, so a retry
removes it again.

### DNS delete

The DNS behavior also supports delete: it removes the interface's own
`/etc/netplan/osapi-dns-<interface>.yaml` and runs `netplan apply`. It returns
`changed: true` if the file existed.

## Safety features

- **Validation with netplan generate.** All writes run `netplan generate` before
  they apply. If the generated configuration is not valid, the agent puts back
  the previous file, and the job returns an error. The previous configuration
  stays.
- **Protection of the default route.** The route behavior does not write a
  default route (`0.0.0.0/0` or `::/0`). Thus a wrong route does not cut the
  host off.
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

## Permissions and platforms

The generated page of each behavior gives the permission, the platforms and the
route of each method: [dns](../reference/behaviors/dns/dns.md),
[interface](../reference/behaviors/interface/interface.md),
[route](../reference/behaviors/route/route.md). The pages come from the
definitions, so they stay true when a method changes, and this page does not
copy them. [Authentication](authentication.md) tells which permissions each role
has.

## CLI usage

```bash
# List all interfaces
osapi client node network interface list --target web-01

# Get a specific interface
osapi client node network interface get \
  --target web-01 --name eth0

# Create a static interface config
osapi client node network interface create \
  --target web-01 --name eth1 \
  --addresses 192.168.1.100/24 \
  --gateway4 192.168.1.1

# Create an interface with DHCP
osapi client node network interface create \
  --target web-01 --name eth1 --dhcp4

# Update an interface (replace config)
osapi client node network interface update \
  --target web-01 --name eth1 \
  --addresses 192.168.1.200/24 \
  --gateway4 192.168.1.1 \
  --mtu 9000

# Delete an interface config
osapi client node network interface delete \
  --target web-01 --name eth1

# List the routes of an interface
osapi client node network interface route list --target web-01 --name eth0

# Add static routes, one a call
osapi client node network interface route create --target web-01 --name eth1 \
  --destination 10.0.0.0/8 --via 192.168.1.1
osapi client node network interface route create --target web-01 --name eth1 \
  --destination 172.16.0.0/12 --via 192.168.1.1 --metric 100

# Change the metric of a route
osapi client node network interface route update --target web-01 --name eth1 \
  --destination 10.0.0.0/8 --via 192.168.1.1 --metric 50

# Drop a route
osapi client node network interface route delete --target web-01 --name eth1 \
  --destination 10.0.0.0/8

# Delete DNS config for an interface
osapi client node network dns delete \
  --target web-01 --interface-name eth0

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

## Container behavior

Interface and route writes return `status: skipped` inside containers. List and
get read the container's own interfaces and routes. Netplan requires direct host
access and is not available in standard container environments.

## Related

- [CLI Reference](../usage/cli/client/node/network/interface/interface.mdx):
  interface commands
- [CLI Reference](../usage/cli/client/node/network/interface/route/route.mdx):
  route commands
- [SDK Reference](../sdk/client/networking/network_interface.md):
  NetworkInterface service
- [SDK Reference](../sdk/client/networking/network_interface_route.md):
  NetworkInterfaceRoute service
- [Network Management](network-management.md): DNS and ping
- [Platform Detection](../sdk/platform/detection.md): OS family detection
- [Configuration](../usage/configuration.md): full configuration reference
