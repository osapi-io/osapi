# Delete

Drop the routes osapi has of an interface.

Removes the interface's routes file and applies what is left. An interface osapi
has no routes of does not change.

```bash
osapi client node network route delete [flags]
```

## Examples

```bash
# An interface osapi has routes of
osapi client node network route delete --target web-01 --interface-name osapi0

# An interface osapi has no routes of
osapi client node network route delete --target web-01 --interface-name osapi0
```

## Result

With the default settings, each example does this on the host:

| Example                          | On the host                                                           |
| -------------------------------- | --------------------------------------------------------------------- |
| An interface osapi has routes of | removes `/etc/netplan/osapi-osapi0-routes.yaml`; runs `netplan apply` |

The command fails in these cases:

| Call                                 | Flags                            | Fails with    |
| ------------------------------------ | -------------------------------- | ------------- |
| A name that is not an interface name | `--interface-name "eth0;reboot"` | InterfaceName |

## Permission

It needs the `network:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Flags

| Flag               | Description                                                                       | Default |
| ------------------ | --------------------------------------------------------------------------------- | ------- |
| `--interface-name` | The interface, such as eth0, or a fact such as @fact.interface.primary (required) |         |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev)                            | `_all`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
