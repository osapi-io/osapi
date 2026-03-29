# Get

Get the node's DNS config:

```bash
$ osapi client node network dns get --interface-name eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  SERVERS                              SEARCH DOMAINS
  server1   192.168.0.247, 2607:f428::1          example.com
```

When targeting all hosts:

```bash
$ osapi client node network dns get --interface-name eth0 --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  SERVERS                      SEARCH DOMAINS
  server1   ok      192.168.0.247, 2607:f428::1  example.com
  server2   ok      8.8.8.8, 1.1.1.1             local
```

When some hosts fail or are skipped, STATUS and ERROR columns are shown:

```bash
$ osapi client node network dns get --interface-name eth0 --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   SERVERS                      SEARCH DOMAINS  ERROR
  server1   ok       192.168.0.247, 2607:f428::1  example.com
  server2   skipped                                               unsupported platform
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
