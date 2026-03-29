# Hostname

Get the node's hostname:

```bash
$ osapi client node hostname get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME
  server1
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

Target by label to query a group of servers:

```bash
$ osapi client node hostname get --target group:web
```

## Update

Set the hostname on the target node:

```bash
$ osapi client node hostname update --name web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED
  web-01    ok      true
```

When targeting all hosts:

```bash
$ osapi client node hostname update --name web-01 --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  ERROR
  server1   ok       true
  server2   skipped           unsupported platform
```

### Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--name`       | New hostname to set (required)                           |         |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
