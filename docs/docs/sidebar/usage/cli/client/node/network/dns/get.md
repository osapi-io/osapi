# Get

Get the DNS resolvers of an interface.

Reads the servers and the search domains of the interface: resolvectl on Linux,
with resolv.conf on a host that has no resolvers per interface, such as a
container, and scutil on macOS.

```bash
osapi client node network dns get [flags]
```

## Examples

```bash
# An interface with servers and a domain
osapi client node network dns get --target web-01 --interface-name eth0

# A container with no resolvers per interface
osapi client node network dns get --target web-01 --interface-name eth0

# A Mac
osapi client node network dns get --target web-01 --interface-name en0
```

## Result

The command fails in these cases:

| Call                                       | Flags                            | Fails with    |
| ------------------------------------------ | -------------------------------- | ------------- |
| An interface the host does not have        | `--interface-name eth9`          | not found     |
| An interface name that is not alphanumeric | `--interface-name "eth0;reboot"` | InterfaceName |

## Permission

It needs the `network:read` permission. `osapi token roles get --name <role>`
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
