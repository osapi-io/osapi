# Get

Get routes for an interface.

```bash
osapi client node network route get
```

## Examples

```bash
osapi client node network route get \
  --target web-01 --interface eth0
osapi client node network route get \
  --target _all --interface eth0
osapi client node network route get \
  --target web-01 --interface eth0 --json
```

## Flags

| Flag          | Description               | Default |
| ------------- | ------------------------- | ------- |
| `--interface` | Interface name (required) |         |

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
