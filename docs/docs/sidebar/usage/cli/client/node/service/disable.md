# Disable

Disable a systemd service from starting on boot.

```bash
osapi client node service disable
```

## Examples

```bash
osapi client node service disable --target web-01 \
  --name nginx.service
osapi client node service disable --target _all \
  --name nginx.service
osapi client node service disable --target web-01 \
  --name nginx.service --json
```

## Result

If the service was enabled, the command returns `changed: true`. If the service
is already disabled, it returns `changed: false`.

## Flags

| Flag     | Description                        | Default |
| -------- | ---------------------------------- | ------- |
| `--name` | Service name to disable (required) |         |

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
