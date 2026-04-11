# Get

Get the node status:

```bash
$ osapi client node status get

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  Hostname: server1              OS: Ubuntu 24.04
  Load: 1.83, 1.96, 2.02 (1m, 5m, 15m)
  Memory: 19 GB used / 31 GB total / 10 GB free

  Disks:
  MOUNT  TOTAL  USED   USAGE
  /      97 GB  56 GB  58%
  /boot  1 GB   0 GB   0%
```

When targeting all hosts, a summary table is shown:

```bash
$ osapi client node status get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  UPTIME                          LOAD  MEM
  server1   ok      64 days, 11 hours, 20 minutes   1.83  19 GB / 31 GB
  server2   ok      12 days, 3 hours, 45 minutes    0.45  8 GB / 16 GB

  2 hosts: 2 ok
```

When some hosts fail or are skipped:

```bash
$ osapi client node status get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  UPTIME                          LOAD  MEM
  server1   ok      64 days, 11 hours, 20 minutes   1.83  19 GB / 31 GB
  server2   skip

  2 hosts: 1 ok, 1 skipped

  Details:
  server2   unsupported platform
```

Target by label to query a group of servers:

```bash
$ osapi client node status get --target group:web
$ osapi client node status get --target group:web.dev
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
