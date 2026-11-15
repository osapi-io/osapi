# Create

Create a group.

```bash
osapi client node group create
```

## Examples

```bash
osapi client node group create --target web-01 --name deploy
```

## Flags

| Flag       | Description                                  | Default |
| ---------- | -------------------------------------------- | ------- |
| `--gid`    | Numeric group ID (system assigns if omitted) |         |
| `--name`   | Group name (required)                        |         |
| `--system` | Create a system group                        |         |

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
