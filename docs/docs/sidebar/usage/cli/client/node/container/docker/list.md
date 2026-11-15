# List

List containers on target node.

```bash
osapi client node container docker list
```

## Examples

```bash
osapi client node container docker list
osapi client node container docker list --state running
osapi client node container docker list --state stopped
```

## Flags

| Flag      | Description                            | Default   |
| --------- | -------------------------------------- | --------- |
| `--limit` | Maximum number of containers to return |           |
| `--state` | Filter by state: running, stopped, all | `running` |

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
