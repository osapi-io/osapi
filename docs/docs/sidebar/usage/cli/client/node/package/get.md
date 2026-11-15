# Get

Get a package by name.

```bash
osapi client node package get
```

## Examples

```bash
osapi client node package get --target web-01 --name nginx
osapi client node package get --target _all --name nginx
osapi client node package get --target web-01 --name nginx --json
```

## Flags

| Flag     | Description                    | Default |
| -------- | ------------------------------ | ------- |
| `--name` | Name of the package (required) |         |

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
