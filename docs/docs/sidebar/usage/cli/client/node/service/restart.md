# Restart

Restart a systemd service.

```bash
osapi client node service restart
```

## Examples

```bash
osapi client node service restart --target web-01 \
  --name nginx.service
osapi client node service restart --target _all \
  --name nginx.service
osapi client node service restart --target web-01 \
  --name nginx.service --json
```

## Flags

| Flag     | Description                        | Default |
| -------- | ---------------------------------- | ------- |
| `--name` | Service name to restart (required) |         |

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
