# Delete

Delete a sysctl parameter.

The agent removes its drop-in file from `/etc/sysctl.d/`. The running kernel
keeps the value until the next reboot, or until you set a new value.

```bash
osapi client node sysctl delete
```

## Examples

```bash
osapi client node sysctl delete --target web-01 \
  --key net.ipv4.ip_forward
osapi client node sysctl delete --target _all \
  --key vm.swappiness
osapi client node sysctl delete --target web-01 \
  --key net.ipv4.ip_forward --json
```

## Result

If osapi does not manage the parameter, the command returns `changed: false`.
The agent does not remove a file that osapi did not write.

## Flags

| Flag    | Description                               | Default |
| ------- | ----------------------------------------- | ------- |
| `--key` | Sysctl parameter key to delete (required) |         |

## Global flags

| Flag               | Description                                            | Default                 |
| ------------------ | ------------------------------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                           |                         |
| `-j, --json`       | Enable JSON output                                     |                         |
| `-f, --osapi-file` | Path to config file                                    | `/etc/osapi/osapi.yaml` |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev) | `_all`                  |
| `--url`            | URL the client will connect to                         | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
