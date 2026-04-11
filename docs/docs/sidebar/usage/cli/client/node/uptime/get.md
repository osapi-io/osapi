# Get

Get uptime from the target node:

```bash
$ osapi client node uptime get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  UPTIME
  web-01    ok      64 days, 11 hours, 20 minutes

  1 host: 1 ok
```

When targeting all hosts:

```bash
$ osapi client node uptime get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  UPTIME
  server1   ok      64 days, 11 hours, 20 minutes
  server2   ok      12 days, 3 hours, 45 minutes

  2 hosts: 2 ok
```

When some hosts fail or are skipped:

```bash
$ osapi client node uptime get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  UPTIME
  server1   ok      64 days, 11 hours, 20 minutes
  server2   skip

  2 hosts: 1 ok, 1 skipped

  Details:
  server2   unsupported platform
```

Target by label to query a group of servers:

```bash
$ osapi client node uptime get --target group:web
$ osapi client node uptime get --target group:web.dev
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
