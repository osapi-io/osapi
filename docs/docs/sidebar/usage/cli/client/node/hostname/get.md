# Get

Get the node's hostname:

```bash
$ osapi client node hostname get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  LABELS
  web-01    group:web
```

When targeting all hosts:

```bash
$ osapi client node hostname get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS
  server1   ok
  server2   ok
```

When some hosts fail or are skipped:

```bash
$ osapi client node hostname get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   ERROR
  server1   ok
  server2   skipped  unsupported platform
```

When a single host does not support the operation:

```bash
$ osapi client node hostname get --target darwin-host

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME     STATUS   ERROR
  darwin-host  skipped  host: operation not supported on this OS family
```

Target by label to query a group of servers:

```bash
$ osapi client node hostname get --target group:web
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
