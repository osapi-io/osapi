# Get

Get uptime from the target node:

```bash
$ osapi client node uptime get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  UPTIME
  64 days, 11 hours, 20 minutes
```

When targeting all hosts:

```bash
$ osapi client node uptime get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  UPTIME
  server1   64 days, 11 hours, 20 minutes
  server2   12 days, 3 hours, 45 minutes
```

When some hosts fail or are skipped:

```bash
$ osapi client node uptime get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   UPTIME                          ERROR
  server1   ok       64 days, 11 hours, 20 minutes
  server2   skipped                                  unsupported platform
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
