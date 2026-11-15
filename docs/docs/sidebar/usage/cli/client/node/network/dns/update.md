# Update

Update the DNS configuration.

```bash
osapi client node network dns update [flags]
```

## Examples

```bash
osapi client node network dns update \
  --servers "1.1.1.1,2.2.2.2" \
  --search-domains "foo.bar,baz.qux" \
  --interface eth0
osapi client node network dns update \
  --servers "1.1.1.1,2.2.2.2" \
  --interface eth0 \
  --target _all
osapi client node network dns update \
  --servers "1.1.1.1,2.2.2.2" \
  --interface eth0 \
  --target group:web
```

## Flags

| Flag               | Description                                                                           | Default |
| ------------------ | ------------------------------------------------------------------------------------- | ------- |
| `--interface`      | Name of the network interface to retrieve DNS server configurations (required)        |         |
| `--override-dhcp`  | When true, disables DHCP-provided DNS servers so only the configured servers are used |         |
| `--search-domains` | List of DNS search domains (comma-separated)                                          |         |
| `--servers`        | List of DNS server IP addresses (comma-separated)                                     |         |

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
