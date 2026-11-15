# Create

Create a sysctl parameter.

The agent writes the value to `/etc/sysctl.d/osapi-{key}.conf` and applies it
immediately with `sysctl -p`. If osapi already manages the parameter, the
command returns `changed: false`.

```bash
osapi client node sysctl create
```

## Examples

```bash
osapi client node sysctl create --target web-01 \
  --key net.ipv4.ip_forward --value 1
osapi client node sysctl create --target _all \
  --key vm.swappiness --value 10
osapi client node sysctl create --target web-01 \
  --key vm.swappiness --value 10 --json
```

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
