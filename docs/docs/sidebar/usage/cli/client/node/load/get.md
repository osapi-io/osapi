# Get

Get load averages from the target node:

```bash
$ osapi client node load get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  LOAD (1m)  LOAD (5m)  LOAD (15m)
  server1   ok      1.83       1.96       2.02
```

When targeting all hosts:

```bash
$ osapi client node load get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  LOAD (1m)  LOAD (5m)  LOAD (15m)
  server1   ok      1.83       1.96       2.02
  server2   ok      0.45       0.52       0.61
```

When some hosts fail or are skipped:

```bash
$ osapi client node load get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   LOAD (1m)  LOAD (5m)  LOAD (15m)  ERROR
  server1   ok       1.83       1.96       2.02
  server2   skipped                                     unsupported platform
```

Target by label to query a group of servers:

```bash
$ osapi client node load get --target group:web
$ osapi client node load get --target group:web.dev
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
