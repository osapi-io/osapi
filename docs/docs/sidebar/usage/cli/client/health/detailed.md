# Detailed

Show per-component health status with version and uptime. This endpoint requires
authentication.

```bash
$ osapi client health detailed

  Status: ok
  Version: 0.1.0
  Uptime: 2h30m15s


  Components:

  ┏━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━┓
  ┃ COMPONENT ┃ STATUS ┃ ERROR ┃
  ┣━━━━━━━━━━━╋━━━━━━━━╋━━━━━━━┫
  ┃ nats      ┃ ok     ┃       ┃
  ┃ kv        ┃ ok     ┃       ┃
  ┗━━━━━━━━━━━┻━━━━━━━━┻━━━━━━━┛
```

When a component is unhealthy, the overall status becomes `degraded` and the
component shows the error:

```bash
$ osapi client health detailed

  Status: degraded
  Version: 0.1.0
  Uptime: 2h30m15s


  Components:

  ┏━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
  ┃ COMPONENT ┃ STATUS ┃ ERROR                       ┃
  ┣━━━━━━━━━━━╋━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
  ┃ nats      ┃ ok     ┃                             ┃
  ┃ kv        ┃ error  ┃ KV bucket not accessible: … ┃
  ┗━━━━━━━━━━━┻━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

Use `--json` for raw JSON output:

```bash
$ osapi client health detailed --json
{"status":"ok","version":"0.1.0","uptime":"2h30m15s","components":{"nats":{"status":"ok"},"kv":{"status":"ok"}}}
```
