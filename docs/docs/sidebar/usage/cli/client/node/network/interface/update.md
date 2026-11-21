# Update

Change the config osapi has of a network interface.

Writes the interface's netplan file over the one osapi has, as create writes it,
and applies it. An interface osapi does not configure is not found: create
configures it.

```bash
osapi client node network interface update [flags]
```

## Examples

```bash
# Move an interface to DHCP
osapi client node network interface update --target web-01 --name osapi0 --dhcp4
```

## Result

With the default settings, each example does this on the host:

| Example                   | On the host                                                                                                                                                                        |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Move an interface to DHCP | runs `netplan generate`; runs `netplan apply`; writes `/etc/netplan/osapi-osapi0.yaml`: `network:\n  version: 2\n  ethernets:\n    osapi0:\n      dhcp4: true\n      dhcp6: false` |

The command fails in these cases:

| Call                                  | Flags                                                                                                                                                                                                                                                                                                                                                                                                                                   | Fails with |
| ------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- |
| An interface osapi does not configure | `--name osapi0 --dhcp4`                                                                                                                                                                                                                                                                                                                                                                                                                 | not found  |
| An address without a prefix           | `--name osapi0 --addresses 10.0.1.5`                                                                                                                                                                                                                                                                                                                                                                                                    | Addresses  |
| More than 16 addresses                | `--name osapi0 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24 --addresses 10.0.1.5/24` | Addresses  |
| An IPv4 gateway that is not IPv4      | `--name osapi0 --gateway4 fe80::1`                                                                                                                                                                                                                                                                                                                                                                                                      | Gateway4   |
| An IPv6 gateway that is not IPv6      | `--name osapi0 --gateway6 10.0.1.1`                                                                                                                                                                                                                                                                                                                                                                                                     | Gateway6   |
| An MTU below 68                       | `--name osapi0 --mtu 10`                                                                                                                                                                                                                                                                                                                                                                                                                | MTU        |
| An MTU above 9000                     | `--name osapi0 --mtu 9001`                                                                                                                                                                                                                                                                                                                                                                                                              | MTU        |
| A MAC address netplan does not take   | `--name osapi0 --mac-address 0000.5e00.5301`                                                                                                                                                                                                                                                                                                                                                                                            | MACAddress |
| A name that is not an interface name  | `--name "eth0;reboot"`                                                                                                                                                                                                                                                                                                                                                                                                                  | Name       |

## Permission

It needs the `network:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag            | Description                                                                       | Default |
| --------------- | --------------------------------------------------------------------------------- | ------- |
| `--addresses`   | The static addresses, in CIDR notation, such as 10.0.0.5/24                       |         |
| `--dhcp4`       | Take an IPv4 address from DHCP                                                    |         |
| `--dhcp6`       | Take an IPv6 address from DHCP                                                    |         |
| `--gateway4`    | The IPv4 default gateway                                                          |         |
| `--gateway6`    | The IPv6 default gateway                                                          |         |
| `--mac-address` | The MAC address to set                                                            |         |
| `--mtu`         | The MTU, from 68 to 9000                                                          |         |
| `--name`        | The interface, such as eth0, or a fact such as @fact.interface.primary (required) |         |
| `-T, --target`  | Target: _all, _any, hostname, or label (group:web.dev)                            | `_all`  |
| `--wakeonlan`   | Turn on wake-on-LAN                                                               |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
