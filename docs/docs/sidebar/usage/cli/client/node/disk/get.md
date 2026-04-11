# Get

Get disk usage from the target node:

```bash
$ osapi client node disk get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  MOUNT  TOTAL   USED    USAGE
  web-01    ok      /      97 GB   56 GB   58%
  web-01    ok      /boot  1 GB    0 GB    0%
  web-01    ok      /home  450 GB  120 GB  27%

  1 host: 1 ok
```

When targeting all hosts:

```bash
$ osapi client node disk get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  MOUNT  TOTAL   USED   USAGE
  server1   ok      /      97 GB   56 GB  58%
  server1   ok      /boot  1 GB    0 GB   0%
  server2   ok      /      200 GB  80 GB  40%

  2 hosts: 2 ok
```

Target by label to query a group of servers:

```bash
$ osapi client node disk get --target group:web
$ osapi client node disk get --target group:web.dev
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
