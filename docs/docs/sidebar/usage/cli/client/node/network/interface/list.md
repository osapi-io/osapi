# List

List the network interfaces of the host.

Reads each interface but the loopback, in the order of its index: /sys/class/net
and ip on Linux, ifconfig and netstat on macOS. dhcp4 is true when a route of
the interface has proto dhcp, and primary when it has the default route.

```bash
osapi client node network interface list [flags]
```

## Examples

```bash
# A Linux host
osapi client node network interface list --target web-01

# A Mac
osapi client node network interface list --target web-01
```

## Permission

It needs the `network:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
