# Delete

Delete a network interface configuration.

```bash
osapi client node network interface delete
```

## Examples

```bash
osapi client node network interface delete \
  --target web-01 --name eth0
osapi client node network interface delete \
  --target _all --name eth1
osapi client node network interface delete \
  --target web-01 --name eth0 --json
```

## Flags

| Flag     | Description                         | Default |
| -------- | ----------------------------------- | ------- |
| `--name` | Interface name to delete (required) |         |

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
