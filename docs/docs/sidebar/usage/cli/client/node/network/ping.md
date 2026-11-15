# Ping

Ping an address from target hosts.

```bash
osapi client node network ping
```

## Examples

```bash
osapi client node network ping --address 8.8.8.8
osapi client node network ping --address 8.8.8.8 --target _all
osapi client node network ping --address 8.8.8.8 --target group:web
```

## Flags

| Flag        | Description                    | Default |
| ----------- | ------------------------------ | ------- |
| `--address` | The address to ping (required) |         |

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
