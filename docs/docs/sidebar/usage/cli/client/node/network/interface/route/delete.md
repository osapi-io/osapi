# Delete

Drop a route that osapi wrote.

Drops the route from the interface's netplan file and applies what is left; the
last route removes the file. A route osapi does not have does not change.

```bash
osapi client node network interface route delete [flags]
```

## Examples

```bash
# One of two routes
osapi client node network interface route delete --target web-01 --name osapi0 --destination 198.51.100.0/24

# The last route
osapi client node network interface route delete --target web-01 --name osapi0 --destination 198.51.100.0/24

# A route osapi does not have
osapi client node network interface route delete --target web-01 --name osapi0 --destination 198.51.100.0/24
```

## Result

With the default settings, each example does this on the host:

| Example           | On the host                                                                                                                                                                                                                                         |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| One of two routes | runs `netplan generate`; runs `netplan apply`; writes `/etc/netplan/osapi-osapi0-routes.yaml`: `network:\n  version: 2\n  ethernets:\n    osapi0:\n      routes:\n        - to: 2001:db8:1::/48\n          via: 2001:db8::1\n          metric: 200` |
| The last route    | removes `/etc/netplan/osapi-osapi0-routes.yaml`; runs `netplan apply`                                                                                                                                                                               |

The command fails in these cases:

| Call                                 | Flags                                                | Fails with  |
| ------------------------------------ | ---------------------------------------------------- | ----------- |
| A destination that is not CIDR       | `--name osapi0 --destination 10.0.0.0`               | Destination |
| A name that is not an interface name | `--name "eth0;reboot" --destination 198.51.100.0/24` | Name        |

## Permission

It needs the `network:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag            | Description                                                                       | Default |
| --------------- | --------------------------------------------------------------------------------- | ------- |
| `--destination` | The destination of the route, in CIDR notation, such as 10.0.0.0/8 (required)     |         |
| `--name`        | The interface, such as eth0, or a fact such as @fact.interface.primary (required) |         |
| `-T, --target`  | Target: _all, _any, hostname, or label (group:web.dev)                            | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../../output.md) tells how the result is printed, and what
`--json` gives.
