# Memory

Get memory usage from the target node:

```bash
$ osapi client node memory get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  MEMORY
  server1   ok      19 GB used / 31 GB total / 10 GB free
```

When targeting all hosts:

```bash
$ osapi client node memory get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  MEMORY
  server1   ok      19 GB used / 31 GB total / 10 GB free
  server2   ok      8 GB used / 16 GB total / 7 GB free
```

When some hosts fail or are skipped:

```bash
$ osapi client node memory get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   MEMORY                               ERROR
  server1   ok       19 GB used / 31 GB total / 10 GB free
  server2   skipped                                       unsupported platform
```

Target by label to query a group of servers:

```bash
$ osapi client node memory get --target group:web
$ osapi client node memory get --target group:web.dev
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
