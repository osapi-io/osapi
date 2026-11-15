# Signal

Send a signal to a process.

```bash
osapi client node process signal
```

## Examples

```bash
osapi client node process signal --target web-01 \
  --pid 1234 --signal TERM
osapi client node process signal --target _all \
  --pid 1234 --signal HUP
osapi client node process signal --target web-01 \
  --pid 1234 --signal TERM --json
```

## Flags

| Flag       | Description                                     | Default |
| ---------- | ----------------------------------------------- | ------- |
| `--pid`    | Process ID to signal (required)                 |         |
| `--signal` | Signal to send, e.g. TERM, KILL, HUP (required) |         |

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
