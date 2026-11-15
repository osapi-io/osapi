# Query

Query journal entries.

```bash
osapi client node log query
```

## Examples

```bash
osapi client node log query --target web-01
osapi client node log query --target web-01 \
  --lines 50 --since 1h --priority err
osapi client node log query --target _all --lines 5
```

## Flags

| Flag         | Description                                                        | Default |
| ------------ | ------------------------------------------------------------------ | ------- |
| `--lines`    | Maximum number of log lines to return                              | `100`   |
| `--priority` | Filter by priority level (e.g., 'err', 'warning', 'info')          |         |
| `--since`    | Return entries since this time (e.g., '1h', '2026-01-01 00:00:00') |         |
| `--unit`     | Only entries from this systemd unit, such as ssh                   |         |

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
