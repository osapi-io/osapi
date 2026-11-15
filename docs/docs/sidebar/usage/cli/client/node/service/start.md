# Start

Start a systemd service.

```bash
osapi client node service start
```

## Examples

```bash
osapi client node service start --target web-01 \
  --name nginx.service
osapi client node service start --target _all \
  --name nginx.service
osapi client node service start --target web-01 \
  --name nginx.service --json
```

## Flags

| Flag     | Description                      | Default |
| -------- | -------------------------------- | ------- |
| `--name` | Service name to start (required) |         |

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
