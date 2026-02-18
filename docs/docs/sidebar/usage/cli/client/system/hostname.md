# Hostname

Get the system's hostname:

```bash
$ osapi client system hostname

  ┏━━━━━━━━━━┓
  ┃ HOSTNAME ┃
  ┣━━━━━━━━━━┫
  ┃ server1  ┃
  ┗━━━━━━━━━━┛
```

When targeting all hosts:

```bash
$ osapi client system hostname --target _all

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
