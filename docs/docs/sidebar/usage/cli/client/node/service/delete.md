# Delete

Delete a systemd service unit file.

If the service runs, the agent stops and disables it. The agent removes the unit
file from `/etc/systemd/system/` and runs `systemctl daemon-reload`.

```bash
osapi client node service delete
```

## Examples

```bash
osapi client node service delete --target web-01 \
  --name myapp.service
osapi client node service delete --target _all \
  --name myapp.service
osapi client node service delete --target web-01 \
  --name myapp.service --json
```

## Result

If the unit file does not exist, the command returns `changed: false`.

## Flags

| Flag     | Description                            | Default |
| -------- | -------------------------------------- | ------- |
| `--name` | Service unit name to delete (required) |         |

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
