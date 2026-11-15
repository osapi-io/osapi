# Update

Update a sysctl parameter.

The agent writes the value again to `/etc/sysctl.d/osapi-{key}.conf` and applies
it immediately with `sysctl -p`.

```bash
osapi client node sysctl update
```

## Examples

```bash
osapi client node sysctl update --target web-01 \
  --key net.ipv4.ip_forward --value 0
osapi client node sysctl update --target _all \
  --key vm.swappiness --value 20
osapi client node sysctl update --target web-01 \
  --key vm.swappiness --value 20 --json
```

## Result

The command is the same as `create`. It sets the parameter, and osapi manages
it, also when osapi did not manage it before. If the file and the kernel have
the requested value, the command returns `changed: false`.

## Flags

| Flag      | Description                       | Default |
| --------- | --------------------------------- | ------- |
| `--key`   | Sysctl parameter key (required)   |         |
| `--value` | Sysctl parameter value (required) |         |

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
