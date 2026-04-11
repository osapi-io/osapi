# Get

Get the node's hostname:

```bash
$ osapi client node hostname get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS
  web-01    ok

  1 host: 1 ok
```

When targeting all hosts:

```bash
$ osapi client node hostname get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS
  server1   ok
  server2   ok

  2 hosts: 2 ok
```

When some hosts fail or are skipped:

```bash
$ osapi client node hostname get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS
  server1   ok
  server2   skip

  2 hosts: 1 ok, 1 skipped

  Details:
  server2   unsupported platform
```

When a single host does not support the operation:

```bash
$ osapi client node hostname get --target darwin-host

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME     STATUS
  darwin-host  skip

  1 host: 1 skipped

  Details:
  darwin-host  host: operation not supported on this OS family
```

Target by label to query a group of servers:

```bash
$ osapi client node hostname get --target group:web
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
