# Ready

Check if the API server is ready to accept traffic. This endpoint verifies that
all dependencies (NATS, KV buckets) are reachable.

```bash
$ osapi client health ready

  Status: ready
```

When the service is not ready:

```bash
$ osapi client health ready

  Status: not_ready
  Error: NATS not connected
```

Use `--json` for raw JSON output:

```bash
$ osapi client health ready --json
{"status":"ready"}
```
