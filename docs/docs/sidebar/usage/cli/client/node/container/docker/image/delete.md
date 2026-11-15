# Delete

Delete a container image.

```bash
osapi client node container docker image delete
```

## Examples

```bash
osapi client node container docker image delete --image nginx:latest
osapi client node container docker image delete \
  --image nginx:latest --force
osapi client node container docker image delete \
  --image redis:7 --target web-01
```

## Flags

| Flag      | Description                           | Default |
| --------- | ------------------------------------- | ------- |
| `--force` | Force removal even if image is in use |         |
| `--image` | Image name or ID to remove (required) |         |

## Global flags

| Flag               | Description                                            | Default                 |
| ------------------ | ------------------------------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                           |                         |
| `-j, --json`       | Enable JSON output                                     |                         |
| `-f, --osapi-file` | Path to config file                                    | `/etc/osapi/osapi.yaml` |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev) | `_all`                  |
| `--url`            | URL the client will connect to                         | `http://0.0.0.0:8080`   |

[Output](../../../../../output.md) tells how the result is printed, and what
`--json` gives.
