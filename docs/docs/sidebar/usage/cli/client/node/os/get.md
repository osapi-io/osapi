# Get

Get operating system information from the target node:

```bash
$ osapi client node os get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  DISTRIBUTION  VERSION
  web-01    ok      Ubuntu        24.04

  1 host: 1 ok
```

When targeting all hosts:

```bash
$ osapi client node os get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  DISTRIBUTION  VERSION
  server1   ok      Ubuntu        24.04
  server2   ok      Debian        12

  2 hosts: 2 ok
```

When some hosts fail or are skipped:

```bash
$ osapi client node os get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  DISTRIBUTION  VERSION
  server1   ok      Ubuntu        24.04
  server2   skip

  2 hosts: 1 ok, 1 skipped

  Details:
  server2   unsupported platform
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
