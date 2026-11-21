# Delete

Drop the config osapi has of a network interface.

Removes the interface's netplan file and applies what is left. An interface
osapi does not configure does not change.

```bash
osapi client node network interface delete [flags]
```

## Examples

```bash
# An interface osapi configures
osapi client node network interface delete --target web-01 --name osapi0

# An interface osapi does not configure
osapi client node network interface delete --target web-01 --name osapi0
```

## Result

With the default settings, each example does this on the host:

| Example                       | On the host                                                    |
| ----------------------------- | -------------------------------------------------------------- |
| An interface osapi configures | removes `/etc/netplan/osapi-osapi0.yaml`; runs `netplan apply` |

The command fails in these cases:

| Call                                 | Flags                  | Fails with |
| ------------------------------------ | ---------------------- | ---------- |
| A name that is not an interface name | `--name "eth0;reboot"` | Name       |

## Permission

It needs the `network:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag           | Description                                                                       | Default |
| -------------- | --------------------------------------------------------------------------------- | ------- |
| `--name`       | The interface, such as eth0, or a fact such as @fact.interface.primary (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev)                            | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
