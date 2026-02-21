# Ping

Ping the desired address:

```bash
$ osapi client network ping --address 8.8.8.8

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  Ping Response:
  ┏━━━━━━━━━━┳━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━┳━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━┓
  ┃ HOSTNAME ┃ AVG RTT      ┃ MAX RTT      ┃ MIN RTT      ┃ PACKET LOSS ┃ PACKETS RECEIVED ┃ PACKETS SENT ┃
  ┣━━━━━━━━━━╋━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━╋━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━┫
  ┃ server1  ┃ 19.707031ms  ┃ 25.066977ms  ┃ 13.007048ms  ┃ 0.000000    ┃ 3                ┃ 3            ┃
  ┗━━━━━━━━━━┻━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━┻━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━┛
```

When targeting all hosts:

```bash
$ osapi client network ping --address 8.8.8.8 --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  Ping Response:
  ┏━━━━━━━━━━┳━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━┳━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━┓
  ┃ HOSTNAME ┃ AVG RTT      ┃ MAX RTT      ┃ MIN RTT      ┃ PACKET LOSS ┃ PACKETS RECEIVED ┃ PACKETS SENT ┃
  ┣━━━━━━━━━━╋━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━╋━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━┫
  ┃ server1  ┃ 19.707031ms  ┃ 25.066977ms  ┃ 13.007048ms  ┃ 0.000000    ┃ 3                ┃ 3            ┃
  ┃ server2  ┃ 22.345678ms  ┃ 28.123456ms  ┃ 18.234567ms  ┃ 0.000000    ┃ 3                ┃ 3            ┃
  ┗━━━━━━━━━━┻━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━┻━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━┛
```

Target by label to ping from a group of servers:

```bash
$ osapi client network ping --address 8.8.8.8 --target group:web
```

## Flags

| Flag            | Description                                              | Default  |
| --------------- | -------------------------------------------------------- | -------- |
| `-a, --address` | The address to ping                                      | required |
| `-T, --target`  | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
