# Create

Create a systemd service unit file.

First, upload the content of the unit file to the Object Store. The agent writes
the file to `/etc/systemd/system/{name}` and runs `systemctl daemon-reload`.

```bash
osapi client node service create
```

## Examples

```bash
osapi client node service create --target web-01 \
  --name myapp.service --object myapp-unit
osapi client node service create --target _all \
  --name myapp.service --object myapp-unit
osapi client node service create --target web-01 \
  --name myapp.service --object myapp-unit --json
```

## Result

If the unit file has the same content, the command returns `changed: false`. If
the file has different content, the agent writes it again.

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
