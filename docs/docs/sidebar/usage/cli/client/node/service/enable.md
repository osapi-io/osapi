# Enable

Enable a systemd service to start on boot.

```bash
osapi client node service enable
```

## Examples

```bash
osapi client node service enable --target web-01 \
  --name nginx.service
osapi client node service enable --target _all \
  --name nginx.service
osapi client node service enable --target web-01 \
  --name nginx.service --json
```

## Result

If the service was not enabled, the command returns `changed: true`. If the
service is already enabled, it returns `changed: false`.

## Flags

| Flag     | Description                       | Default |
| -------- | --------------------------------- | ------- |
| `--name` | Service name to enable (required) |         |

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
