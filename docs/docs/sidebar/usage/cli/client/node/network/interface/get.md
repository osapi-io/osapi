# Get

Get a network interface of the host.

Reads the interface as list does.

```bash
osapi client node network interface get [flags]
```

## Examples

```bash
# An interface with a DHCP address
osapi client node network interface get --target web-01 --name eth0

# A Mac
osapi client node network interface get --target web-01 --name en0
```

## Result

The command fails in these cases:

| Call                                 | Flags                      | Fails with |
| ------------------------------------ | -------------------------- | ---------- |
| An interface the host does not have  | `--name eth9`              | not found  |
| A name that is not an interface name | `--name "eth0;reboot"`     | Name       |
| A fact the agent does not know       | `--name @fact.nonexistent` | Name       |

## Permission

It needs the `network:read` permission. `osapi token roles get --name <role>`
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
