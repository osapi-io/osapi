# Status

Show the system status. The output has: a table of all components, the health of
each component, the NATS connection, statistics for the streams and the KV
buckets, the consumers and the job queue counts. This endpoint needs
authentication.

```bash
$ osapi client health status

  Status: ok
  Version: 0.1.0
  Uptime: 7h6m

  TYPE        HOSTNAME  STATUS  CONDITIONS    AGE     CPU    MEM
  agent       web-01    Ready   DiskPressure  7h 6m   1.2%   96 MB
              ├─ heartbeat               ok
              └─ metrics                 ok http://0.0.0.0:9091
  agent       web-02    Ready   -             3h 2m   0.8%   82 MB
              ├─ heartbeat               ok
              └─ metrics                 ok http://0.0.0.0:9091
  controller  web-01    Ready   -             7h 6m   2.1%   128 MB
              ├─ api                     ok http://0.0.0.0:8080
              ├─ heartbeat               ok
              ├─ kv (connectivity)        ok
              ├─ metrics                 ok http://0.0.0.0:9090
              ├─ nats (connectivity)      ok
              ├─ notifier                ok
              └─ tracing                 ok
  nats        web-01    Ready   -             7h 6m   0.3%   64 MB
              ├─ heartbeat               ok
              ├─ metrics                 ok http://0.0.0.0:9092
              └─ server                  ok nats://localhost:4222

  Consumers: 24 total
  Jobs: 100 total, 90 completed, 5 unprocessed, 3 failed, 0 dlq
  Stream: JOBS (42 msgs, 1.0 KB, 24 consumers)
  Bucket: job-queue (10 keys, 2.0 KB)
  Bucket: agent-registry (4 keys, 1.2 KB)
  Bucket: agent-facts (2 keys, 1.5 KB)
  Bucket: audit-log (50 keys, 8.0 KB)
```

The component table shows all registered components with their health status,
active conditions, uptime, and process resource usage. Sub-components are
displayed nested under their parent with addresses where applicable.
Connectivity checks show the controller's ability to reach NATS and KV
dependencies. Use `osapi client agent list` for agent-specific details like
labels and OS info.

When a component is not healthy, the overall status is `degraded`, and the
component shows the error:

```bash
$ osapi client health status

  Status: degraded
  Version: 0.1.0
  Uptime: 7h6m

  TYPE        HOSTNAME  STATUS  CONDITIONS  AGE     CPU    MEM
  controller  web-01    Ready   -           7h 6m   2.1%   128 MB
              ├─ api                     ok http://0.0.0.0:8080
              ├─ heartbeat               ok
              ├─ kv (connectivity)        error KV bucket not accessible
              ├─ metrics                 ok http://0.0.0.0:9090
              ├─ nats (connectivity)      ok
              ├─ notifier                disabled
              └─ tracing                 ok
  agent       web-01    Ready   -           7h 6m   1.2%   96 MB
              ├─ heartbeat               ok
              └─ metrics                 ok http://0.0.0.0:9091
```

Use `--json` to get the raw JSON output:

```bash
$ osapi client health status --json
{"status":"ok","version":"0.1.0",...}
```
