# Get

Get a sysctl parameter by key.

```bash
osapi client node sysctl get
```

## Examples

```bash
osapi client node sysctl get --target web-01 --key net.ipv4.ip_forward
osapi client node sysctl get --target _all --key vm.swappiness
osapi client node sysctl get --target web-01 --key vm.swappiness --json
```

## Flags

| Flag    | Description                     | Default |
| ------- | ------------------------------- | ------- |
| `--key` | Sysctl parameter key (required) |         |

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
