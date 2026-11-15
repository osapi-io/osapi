# Shutdown

Shut down the target host.

The agent runs `shutdown -h` with the configured delay. When `--delay` is 0, the
host shuts down immediately.

```bash
osapi client node power shutdown
```

## Examples

```bash
osapi client node power shutdown --target web-01
osapi client node power shutdown --target web-01 \
  --delay 60 --message "Scheduled maintenance shutdown"
osapi client node power shutdown --target _all --delay 30
```

## Flags

| Flag        | Description                                      | Default |
| ----------- | ------------------------------------------------ | ------- |
| `--delay`   | Seconds to wait before shutting down (default 0) |         |
| `--message` | Optional message to broadcast before shutdown    |         |

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
