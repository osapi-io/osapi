# Pull

Pull a container image.

```bash
osapi client node container docker pull
```

## Examples

```bash
osapi client node container docker pull --image nginx:latest
osapi client node container docker pull --image alpine:3.18
osapi client node container docker pull \
  --image registry.example.com/myapp:v1.2.3
```

## Flags

| Flag      | Description                        | Default |
| --------- | ---------------------------------- | ------- |
| `--image` | Image reference to pull (required) |         |

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
