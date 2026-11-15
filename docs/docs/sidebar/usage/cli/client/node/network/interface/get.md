# Get

Get a network interface.

```bash
osapi client node network interface get
```

## Examples

```bash
osapi client node network interface get \
  --target web-01 --name eth0
osapi client node network interface get \
  --target _all --name eth0
osapi client node network interface get \
  --target web-01 --name eth0 --json
```

## Flags

| Flag     | Description               | Default |
| -------- | ------------------------- | ------- |
| `--name` | Interface name (required) |         |

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
