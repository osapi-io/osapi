# Get

Get disk usage from the target node:

```bash
$ osapi client node disk get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  Hostname: server1

  DISK NAME  TOTAL   USED   FREE
  /          97 GB   56 GB  36 GB
  /boot      1 GB    0 GB   1 GB
  /home      450 GB  120 GB 310 GB
```

When targeting all hosts, each host's disks are listed separately:

```bash
$ osapi client node disk get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  Hostname: server1

  DISK NAME  TOTAL   USED   FREE
  /          97 GB   56 GB  36 GB
  /boot      1 GB    0 GB   1 GB

  Hostname: server2

  DISK NAME  TOTAL    USED    FREE
  /          200 GB   80 GB   110 GB
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
