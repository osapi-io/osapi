# Update

Update the node's DNS config:

```bash
$ osapi client node network dns update \
    --servers "1.1.1.1,2.2.2.2" \
    --search-domains "foo.bar,baz.qux" \
    --interface eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true

  1 host: 1 changed
```

When you target all hosts:

```bash
$ osapi client node network dns update \
    --servers "1.1.1.1,2.2.2.2" \
    --interface eth0 \
    --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  server1   changed  true
  server2   skip

  2 hosts: 1 changed, 1 skipped

  Details:
  server2   unsupported platform
```

Target a label to update a group of servers:

```bash
$ osapi client node network dns update \
    --servers "1.1.1.1,2.2.2.2" \
    --interface eth0 \
    --target group:web
```

Replace the DNS servers from DHCP, so that the host uses only the configured
servers:

```bash
$ osapi client node network dns update \
    --servers "1.1.1.1,2.2.2.2" \
    --interface eth0 \
    --override-dhcp
```

## Flags

| Flag               | Description                                                   | Default  |
| ------------------ | ------------------------------------------------------------- | -------- |
| `--servers`        | List of DNS server IP addresses                               | one of\* |
| `--search-domains` | List of DNS search domains                                    | one of\* |
| `--interface`      | Name of the network interface to configure DNS                | required |
| `--override-dhcp`  | Disable DHCP-provided DNS servers, use only configured values | `false`  |
| `-T, --target`     | Target: `_any`, `_all`, hostname, or label (`group:web`)      | `_all`   |

\*Give `--servers`, `--search-domains`, or both.
