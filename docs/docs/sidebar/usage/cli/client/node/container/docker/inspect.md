# Inspect

Inspect a container.

```bash
osapi client node container docker inspect
```

## Examples

```bash
osapi client node container docker inspect --id a1b2c3d4e5f6
osapi client node container docker inspect --id my-nginx
osapi client node container docker inspect --id my-nginx --target web-01
```

## Flags

| Flag   | Description                        | Default |
| ------ | ---------------------------------- | ------- |
| `--id` | Container ID to inspect (required) |         |

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
