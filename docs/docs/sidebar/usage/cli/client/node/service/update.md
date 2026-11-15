# Update

Update a systemd service unit file.

The agent deploys the unit file again and runs `systemctl daemon-reload`.

```bash
osapi client node service update
```

## Examples

```bash
osapi client node service update --target web-01 \
  --name myapp.service --object myapp-unit-v2
osapi client node service update --target _all \
  --name myapp.service --object myapp-unit-v2
osapi client node service update --target web-01 \
  --name myapp.service --object myapp-unit-v2 --json
```

## Result

The command fails if the unit file does not exist. If the content did not change
(the same SHA), the command returns `changed: false`.

## Flags

| Flag       | Description                                         | Default |
| ---------- | --------------------------------------------------- | ------- |
| `--name`   | Service unit name (required)                        |         |
| `--object` | Object store reference for the unit file (required) |         |

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
