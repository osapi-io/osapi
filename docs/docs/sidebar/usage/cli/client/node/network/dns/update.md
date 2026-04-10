# Update

Update the node's DNS config:

```bash
$ osapi client node network dns update \
    --servers "1.1.1.1,2.2.2.2" \
    --search-domains "foo.bar,baz.qux" \
    --interface-name eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED  ERROR
  web-01    ok      true
```

When targeting all hosts, HOSTNAME is shown. STATUS and ERROR columns appear
when any host has an error or is skipped:

```bash
$ osapi client node network dns update \
    --servers "1.1.1.1,2.2.2.2" \
    --interface-name eth0 \
    --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  ERROR
  server1   ok       true
  server2   skipped           unsupported platform
```

Target by label to update a group of servers:

```bash
$ osapi client node network dns update \
    --servers "1.1.1.1,2.2.2.2" \
    --interface-name eth0 \
    --target group:web
```

Override DHCP-provided DNS servers so only the configured servers are used:

```bash
$ osapi client node network dns update \
    --servers "1.1.1.1,2.2.2.2" \
    --interface-name eth0 \
    --override-dhcp
```

## Flags

| Flag               | Description                                                   | Default  |
| ------------------ | ------------------------------------------------------------- | -------- |
| `--servers`        | List of DNS server IP addresses                               | one of\* |
| `--search-domains` | List of DNS search domains                                    | one of\* |
| `--interface-name` | Name of the network interface to configure DNS                | required |
| `--override-dhcp`  | Disable DHCP-provided DNS servers, use only configured values | `false`  |
| `-T, --target`     | Target: `_any`, `_all`, hostname, or label (`group:web`)      | `_all`   |

\*At least one of `--servers` or `--search-domains` must be provided.
