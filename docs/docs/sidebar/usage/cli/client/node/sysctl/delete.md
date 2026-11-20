# Delete

Stop the management of a kernel parameter.

Removes the file that osapi wrote for it and reloads with sysctl --system. Then
a setting for the key in a different file takes effect. If there is no such
setting, the parameter keeps its current value until the next reboot. A delete
of a parameter that osapi does not manage reports changed: false.

```bash
osapi client node sysctl delete [flags]
```

## Examples

```bash
# Remove a parameter
osapi client node sysctl delete --target web-01 --key vm.swappiness

# A parameter osapi does not manage
osapi client node sysctl delete --target web-01 --key vm.swappiness

# A parameter osapi already stopped managing
osapi client node sysctl delete --target web-01 --key vm.swappiness
```

## Result

With the default settings, each example does this on the host:

| Example            | On the host                                                              |
| ------------------ | ------------------------------------------------------------------------ |
| Remove a parameter | removes `/etc/sysctl.d/osapi-vm.swappiness.conf`; runs `sysctl --system` |

The command fails in these cases:

| Call                               | Flags                          | Fails with            |
| ---------------------------------- | ------------------------------ | --------------------- |
| No key                             |                                | key must not be empty |
| A key that is not a parameter name | `--key "vm.swappiness;reboot"` | invalid sysctl key    |
| A key holding a value              | `--key vm.swappiness=1`        | invalid sysctl key    |

## Permission

It needs the `sysctl:write` permission. `osapi token roles get --name <role>`
shows what a role grants.

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

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
