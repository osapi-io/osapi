# Ping

Ping the desired address from the target node:

```bash
$ osapi client node network ping --address 8.8.8.8

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  AVG          MIN          MAX          LOSS
  server1   ok      19.707031ms  13.007048ms  25.066977ms  0.000000

  1 host: 1 ok
```

When targeting all hosts:

```bash
$ osapi client node network ping --address 8.8.8.8 --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  AVG          MIN          MAX          LOSS
  server1   ok      19.707031ms  13.007048ms  25.066977ms  0.000000
  server2   ok      22.345678ms  18.234567ms  28.123456ms  0.000000

  2 hosts: 2 ok
```

When some hosts fail or are skipped:

```bash
$ osapi client node network ping --address 8.8.8.8 --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  AVG          MIN          MAX          LOSS
  server1   ok      19.707031ms  13.007048ms  25.066977ms  0.000000
  server2   skip

  2 hosts: 1 ok, 1 skipped

  Details:
  server2   unsupported platform
```

Target by label to ping from a group of servers:

```bash
$ osapi client node network ping --address 8.8.8.8 --target group:web
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--address`    | The address to ping                                      | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
