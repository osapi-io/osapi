# Create

Create a network interface configuration.

The command fails if osapi already manages a configuration for that interface.

```bash
osapi client node network interface create
```

## Examples

```bash
osapi client node network interface create \
  --target web-01 --name eth0 \
  --address 192.168.1.100/24 \
  --gateway4 192.168.1.1
osapi client node network interface create \
  --target web-01 --name eth1 --dhcp4
osapi client node network interface create \
  --target _all --name eth1 --dhcp4
```

## Flags

| Flag            | Description                              | Default |
| --------------- | ---------------------------------------- | ------- |
| `--address`     | IP address in CIDR notation (repeatable) |         |
| `--dhcp4`       | Enable DHCPv4                            |         |
| `--dhcp6`       | Enable DHCPv6                            |         |
| `--gateway4`    | IPv4 gateway address                     |         |
| `--gateway6`    | IPv6 gateway address                     |         |
| `--mac-address` | Hardware MAC address                     |         |
| `--mtu`         | Maximum transmission unit                |         |
| `--name`        | Interface name (required)                |         |
| `--wakeonlan`   | Enable Wake-on-LAN                       |         |

## Global flags

| Flag               | Description                                            | Default                 |
| ------------------ | ------------------------------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                           |                         |
| `-j, --json`       | Enable JSON output                                     |                         |
| `-f, --osapi-file` | Path to config file                                    | `/etc/osapi/osapi.yaml` |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev) | `_all`                  |
| `--url`            | URL the client will connect to                         | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
