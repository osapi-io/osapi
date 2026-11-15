# Create

Create a new container.

```bash
osapi client node container docker create
```

## Examples

```bash
osapi client node container docker create --image nginx:latest
osapi client node container docker create \
  --image nginx:latest \
  --name my-nginx \
  --env "PORT=8080" --env "DEBUG=true" \
  --port "8080:80" --port "8443:443" \
  --volume "/data:/var/lib/data"
osapi client node container docker create \
  --image alpine:latest \
  --name my-alpine \
  --auto-start=false
```

## Flags

| Flag           | Description                                           | Default |
| -------------- | ----------------------------------------------------- | ------- |
| `--auto-start` | Start the container immediately after creation        | `true`  |
| `--env`        | Environment variable in KEY=VALUE format (repeatable) |         |
| `--image`      | Container image reference (required)                  |         |
| `--name`       | Optional name for the container                       |         |
| `--port`       | Port mapping in host:container format (repeatable)    |         |
| `--volume`     | Volume mount in host:container format (repeatable)    |         |

## Global flags

| Flag               | Description                                            | Default                 |
| ------------------ | ------------------------------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                           |                         |
| `-j, --json`       | Enable JSON output                                     |                         |
| `-f, --osapi-file` | Path to config file                                    | `/etc/osapi/osapi.yaml` |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev) | `_all`                  |
| `--url`            | URL the client will connect to                         | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
