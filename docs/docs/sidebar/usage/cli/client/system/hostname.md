# Hostname

Get the system's hostname:

```bash
$ osapi client system hostname

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  ┏━━━━━━━━━━┓
  ┃ HOSTNAME ┃
  ┣━━━━━━━━━━┫
  ┃ server1  ┃
  ┗━━━━━━━━━━┛
```

When targeting all hosts:

```bash
$ osapi client system hostname --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  ┏━━━━━━━━━━┓
  ┃ HOSTNAME ┃
  ┣━━━━━━━━━━┫
  ┃ server1  ┃
  ┃ server2  ┃
  ┗━━━━━━━━━━┛
```

Target by label to query a group of servers:

```bash
$ osapi client system hostname --target group:web
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
