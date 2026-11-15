# List

List installed packages, or with --updates those with an update.

```bash
osapi client node package list [flags]
```

## Examples

```bash
osapi client node package list --target web-01
osapi client node package list --target _all
osapi client node package list --target group:web
```

## Flags

| Flag        | Description                                 | Default |
| ----------- | ------------------------------------------- | ------- |
| `--updates` | List only packages with an update available |         |

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
