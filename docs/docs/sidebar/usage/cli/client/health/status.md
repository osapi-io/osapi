# Status

Show system status with per-component health, NATS connection info, stream and
KV bucket statistics, and job queue counts. This endpoint requires
authentication.

```bash
$ osapi client health status

  Status: ok              Version: 0.1.0          Uptime: 2h30m
  NATS: nats://localhost:4222 (v2.12.4)
  Jobs: 100 total, 90 completed, 5 unprocessed, 3 failed, 0 dlq

  Components:

  ┏━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━┓
  ┃ COMPONENT ┃ STATUS ┃ ERROR ┃
  ┣━━━━━━━━━━━╋━━━━━━━━╋━━━━━━━┫
  ┃ nats      ┃ ok     ┃       ┃
  ┃ kv        ┃ ok     ┃       ┃
  ┗━━━━━━━━━━━┻━━━━━━━━┻━━━━━━━┛

  Streams:

  ┏━━━━━━┳━━━━━━━━━━┳━━━━━━━┳━━━━━━━━━━━┓
  ┃ NAME ┃ MESSAGES ┃ BYTES ┃ CONSUMERS ┃
  ┣━━━━━━╋━━━━━━━━━━╋━━━━━━━╋━━━━━━━━━━━┫
  ┃ JOBS ┃ 42       ┃ 1024  ┃ 1         ┃
  ┗━━━━━━┻━━━━━━━━━━┻━━━━━━━┻━━━━━━━━━━━┛

  KV Buckets:

  ┏━━━━━━━━━━━┳━━━━━━┳━━━━━━━┓
  ┃ NAME      ┃ KEYS ┃ BYTES ┃
  ┣━━━━━━━━━━━╋━━━━━━╋━━━━━━━┫
  ┃ job-queue ┃ 10   ┃ 2048  ┃
  ┗━━━━━━━━━━━┻━━━━━━┻━━━━━━━┛
```

When a component is unhealthy, the overall status becomes `degraded` and the
component shows the error:

```bash
$ osapi client health status

  Status: degraded        Version: 0.1.0          Uptime: 2h30m
  NATS: nats://localhost:4222 (v2.12.4)
  Jobs: 100 total, 90 completed, 5 unprocessed, 3 failed, 0 dlq

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
$ osapi client health status --json
{"status":"ok","version":"0.1.0","uptime":"2h30m","components":{"nats":{"status":"ok"},"kv":{"status":"ok"}},"nats":{"url":"nats://localhost:4222","version":"2.12.4"},"streams":[{"name":"JOBS","messages":42,"bytes":1024,"consumers":1}],"kv_buckets":[{"name":"job-queue","keys":10,"bytes":2048}],"jobs":{"total":100,"unprocessed":5,"processing":2,"completed":90,"failed":3,"dlq":0}}
```
