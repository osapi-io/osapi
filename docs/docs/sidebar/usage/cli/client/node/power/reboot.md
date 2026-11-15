# Reboot

Reboot the target host.

The agent runs `shutdown -r` with the configured delay. When `--delay` is 0, the
host reboots immediately.

```bash
osapi client node power reboot
```

## Examples

```bash
osapi client node power reboot --target web-01
osapi client node power reboot --target web-01 \
  --delay 60 --message "Scheduled maintenance reboot"
osapi client node power reboot --target _all --delay 30
```

## Flags

| Flag        | Description                                  | Default |
| ----------- | -------------------------------------------- | ------- |
| `--delay`   | Seconds to wait before rebooting (default 0) |         |
| `--message` | Optional message to broadcast before reboot  |         |

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
