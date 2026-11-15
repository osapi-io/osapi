# Update

Update the node's hostname.

```bash
osapi client node hostname update [flags]
```

## Examples

```bash
osapi client node hostname update --name web-01
osapi client node hostname update --name web-01 --target darwin-host
osapi client node hostname update --name web-01 --target _all
```

## Flags

| Flag     | Description                    | Default |
| -------- | ------------------------------ | ------- |
| `--name` | New hostname to set (required) |         |

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
