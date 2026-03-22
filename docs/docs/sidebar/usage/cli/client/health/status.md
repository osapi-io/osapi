# Status

Show system status with a unified component table, per-component health, NATS
connection info, stream and KV bucket statistics, consumer details, and job
queue counts. This endpoint requires authentication.

```bash
$ osapi client health status

  TYPE        HOSTNAME  STATUS  CONDITIONS    AGE     CPU    MEM
  agent       web-01    Ready   DiskPressure  7h 6m   1.2%   96 MB
              ├─ heartbeat               ok
              └─ metrics                 ok
  agent       web-02    Ready   -             3h 2m   0.8%   82 MB
              ├─ heartbeat               ok
              └─ metrics                 ok
  controller  web-01    Ready   -             7h 6m   2.1%   128 MB
              ├─ heartbeat               ok
              ├─ metrics                 ok
              ├─ notifier                ok
              └─ tracing                 ok
  nats        web-01    Ready   -             7h 6m   0.3%   64 MB
              ├─ heartbeat               ok
              └─ metrics                 ok

  Status: ok              Version: 0.1.0          Uptime: 7h6m
  NATS: ok nats://localhost:4222 (v2.12.5)
  KV: ok
  Consumers: 24 total
  Jobs: 100 total, 90 completed, 5 unprocessed, 3 failed, 0 dlq
  Stream: JOBS (42 msgs, 1.0 KB, 24 consumers)
  Bucket: job-queue (10 keys, 2.0 KB)
  Bucket: agent-registry (4 keys, 1.2 KB)
  Bucket: agent-facts (2 keys, 1.5 KB)
  Bucket: audit-log (50 keys, 8.0 KB)
```

The component table shows all registered components with their health status,
active conditions, uptime, and process resource usage. Sub-components
(heartbeat, metrics, notifier, tracing) are displayed nested under their parent.
Use `osapi client agent list` for agent-specific details like labels and OS
info.

When a component is unhealthy, the overall status becomes `degraded` and the
component shows the error:

```bash
$ osapi client health status

  TYPE        HOSTNAME  STATUS  CONDITIONS  AGE     CPU    MEM
  controller  web-01    Ready   -           7h 6m   2.1%   128 MB
              ├─ heartbeat               ok
              ├─ metrics                 ok
              ├─ notifier                disabled
              └─ tracing                 ok
  agent       web-01    Ready   -           7h 6m   1.2%   96 MB
              ├─ heartbeat               ok
              └─ metrics                 ok

  Status: degraded        Version: 0.1.0          Uptime: 7h6m
  NATS: ok nats://localhost:4222 (v2.12.5)
  KV: error KV bucket not accessible
```

Use `--json` for raw JSON output:

```bash
$ osapi client health status --json
{"status":"ok","version":"0.1.0",...}
```
