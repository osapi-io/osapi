# Status

Show system status with per-component health, NATS connection info, stream and
KV bucket statistics, agent details, consumer details, and job queue counts.
This endpoint requires authentication.

```bash
$ osapi client health status

  Status: ok              Version: 0.1.0          Uptime: 2h30m
  NATS: ok nats://localhost:4222 (v2.12.4)
  KV: ok
  Agents: 2 total, 2 ready
  HOSTNAME  LABELS          REGISTERED
  web-01    group=web.prod  15s ago
  web-02    group=web.prod  8s ago
  Consumers: 24 total
  NAME                 PENDING  ACK PENDING  REDELIVERED
  query_any_web_01     0        0            0
  modify_any_web_01    0        0            0
  Jobs: 100 total, 90 completed, 5 unprocessed, 3 failed, 0 dlq
  Stream: JOBS (42 msgs, 1.0 KB, 24 consumers)
  Bucket: job-queue (10 keys, 2.0 KB)
  Bucket: agent-registry (2 keys, 256 B)
  Bucket: audit-log (50 keys, 8.0 KB)
```

When a component is unhealthy, the overall status becomes `degraded` and the
component shows the error:

```bash
$ osapi client health status

  Status: degraded        Version: 0.1.0          Uptime: 2h30m
  NATS: ok nats://localhost:4222 (v2.12.4)
  KV: error KV bucket not accessible
  Jobs: 100 total, 90 completed, 5 unprocessed, 3 failed, 0 dlq
```

Use `--json` for raw JSON output:

```bash
$ osapi client health status --json
{"status":"ok","version":"0.1.0",...}
```
