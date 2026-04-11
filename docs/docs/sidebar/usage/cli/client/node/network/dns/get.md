# Get

Get the node's DNS config:

```bash
$ osapi client node network dns get --interface-name eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  SERVERS                              SEARCH DOMAINS
  web-01    ok      192.168.0.247, 2607:f428::1          example.com

  1 host: 1 ok
```

When targeting all hosts:

```bash
$ osapi client node network dns get --interface-name eth0 --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  SERVERS                      SEARCH DOMAINS
  server1   ok      192.168.0.247, 2607:f428::1  example.com
  server2   ok      8.8.8.8, 1.1.1.1             local

  2 hosts: 2 ok
```

When some hosts fail or are skipped:

```bash
$ osapi client node network dns get --interface-name eth0 --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  SERVERS                      SEARCH DOMAINS
  server1   ok      192.168.0.247, 2607:f428::1  example.com
  server2   skip

  2 hosts: 1 ok, 1 skipped

  Details:
  server2   unsupported platform
```

Target by label to query a group of servers:

```bash
$ osapi client node network dns get --interface-name eth0 --target group:web
```

## Flags

| Flag               | Description                                              | Default  |
| ------------------ | -------------------------------------------------------- | -------- |
| `--interface-name` | Name of the network interface to query DNS for           | required |
| `-T, --target`     | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
