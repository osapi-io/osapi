# Get

Get a kernel parameter's current value, and whether osapi manages it.

```bash
osapi client next node sysctl get [flags]
```

## Examples

```bash
# A parameter osapi manages
osapi client next node sysctl get --target web-01 --key vm.swappiness

# A parameter osapi does not manage
osapi client next node sysctl get --target web-01 --key vm.swappiness

# A value of several fields
osapi client next node sysctl get --target web-01 --key net.ipv4.ping_group_range

# A managed parameter the kernel does not have right now
osapi client next node sysctl get --target web-01 --key net.bridge.bridge-nf-call-iptables
```

## Result

The command fails in these cases:

| Call                                 | Flags                          | Fails with            |
| ------------------------------------ | ------------------------------ | --------------------- |
| A parameter the kernel does not have | `--key vm.nosuch`              | not found             |
| No key                               |                                | key must not be empty |
| A key that is not a parameter name   | `--key "vm.swappiness;reboot"` | invalid sysctl key    |
| A key holding a value                | `--key vm.swappiness=1`        | invalid sysctl key    |

## Permission

It needs the `sysctl:read` permission. `osapi token roles list` lists the roles
that grant it.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--key`        | Parameter name, such as net.ipv4.ip_forward (required) |         |
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
