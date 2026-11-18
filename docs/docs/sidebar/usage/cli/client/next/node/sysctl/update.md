# Update

Set a kernel parameter, manage it, and apply it.

The same as create, addressed by the parameter: osapi manages the parameter from
then on whether it did before or not. Setting the value it already has reports
changed: false.

```bash
osapi client next node sysctl update [flags]
```

## Examples

```bash
# Lower swappiness
osapi client next node sysctl update --target web-01 --key vm.swappiness --value 10

# A parameter osapi did not manage
osapi client next node sysctl update --target web-01 --key vm.swappiness --value 10
```

## Result

With the default settings, each example does this on the host:

| Example                          | On the host                                                                                                                    |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| Lower swappiness                 | writes `/etc/sysctl.d/osapi-vm.swappiness.conf`: `vm.swappiness = 10`; runs `sysctl -p /etc/sysctl.d/osapi-vm.swappiness.conf` |
| A parameter osapi did not manage | writes `/etc/sysctl.d/osapi-vm.swappiness.conf`: `vm.swappiness = 10`; runs `sysctl -p /etc/sysctl.d/osapi-vm.swappiness.conf` |

The command fails in these cases:

| Call                               | Flags                                                | Fails with              |
| ---------------------------------- | ---------------------------------------------------- | ----------------------- |
| No key                             | `--value 10`                                         | key must not be empty   |
| A key that is not a parameter name | `--key "vm.swappiness;reboot" --value 1`             | invalid sysctl key      |
| A key holding a value              | `--key vm.swappiness=1 --value 1`                    | invalid sysctl key      |
| No value                           | `--key vm.swappiness`                                | value must not be empty |
| A value on two lines               | `--key vm.swappiness --value "10\nkernel.panic = 0"` | line breaks             |

## Permission

It needs the `sysctl:write` permission. `osapi token roles` lists the roles that
grant it.

## Flags

| Flag           | Description                                            | Default |
| -------------- | ------------------------------------------------------ | ------- |
| `--key`        | Parameter name, such as net.ipv4.ip_forward (required) |         |
| `-T, --target` | Target: _all, _any, hostname, or label (group:web.dev) | `_all`  |
| `--value`      | Value to set (required)                                |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
