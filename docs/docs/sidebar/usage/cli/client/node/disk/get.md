# Get

Get disk usage from the target node:

```bash
$ osapi client node disk get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  MOUNT  TOTAL   USED    FREE    USAGE
  /      97 GB   56 GB   36 GB   58%
  /boot  1 GB    0 GB    1 GB    0%
  /home  450 GB  120 GB  310 GB  27%
```

When targeting all hosts, the HOSTNAME column is shown:

```bash
$ osapi client node disk get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  MOUNT  TOTAL   USED   FREE    USAGE
  server1   /      97 GB   56 GB  36 GB   58%
  server1   /boot  1 GB    0 GB   1 GB    0%
  server2   /      200 GB  80 GB  110 GB  40%
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
