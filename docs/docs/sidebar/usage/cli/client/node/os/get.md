# Get

Get operating system information from the target node:

```bash
$ osapi client node os get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  DISTRIBUTION  VERSION
  server1   ok      Ubuntu        24.04
```

When targeting all hosts:

```bash
$ osapi client node os get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  DISTRIBUTION  VERSION
  server1   ok      Ubuntu        24.04
  server2   ok      Debian        12
```

When some hosts fail or are skipped:

```bash
$ osapi client node os get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   DISTRIBUTION  VERSION  ERROR
  server1   ok       Ubuntu        24.04
  server2   skipped                          unsupported platform
```

Target by label to query a group of servers:

```bash
$ osapi client node os get --target group:web
$ osapi client node os get --target group:web.dev
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
