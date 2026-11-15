# Get

Get process information by PID.

```bash
osapi client node process get
```

## Examples

```bash
osapi client node process get --target web-01 --pid 1234
osapi client node process get --target _all --pid 1
osapi client node process get --target web-01 --pid 1234 --json
```

## Flags

| Flag    | Description                      | Default |
| ------- | -------------------------------- | ------- |
| `--pid` | Process ID to inspect (required) |         |

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
