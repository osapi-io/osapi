# Delete

Delete the DNS configuration.

The agent removes the `/etc/netplan/osapi-dns.yaml` file and runs
`netplan apply`.

```bash
osapi client node network dns delete
```

## Examples

```bash
osapi client node network dns delete \
  --interface eth0
osapi client node network dns delete \
  --interface eth0 \
  --target _all
osapi client node network dns delete \
  --interface eth0 --json
```

## Result

If osapi does not manage a DNS configuration, the command returns
`changed: false`.

## Flags

| Flag          | Description                              | Default |
| ------------- | ---------------------------------------- | ------- |
| `--interface` | Name of the network interface (required) |         |

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
