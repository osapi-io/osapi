# Ready

Check if the API server is ready for traffic. This endpoint checks that it can
reach all its dependencies, NATS and the KV buckets.

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

Use `--json` to get the raw JSON output:

```bash
$ osapi client health ready --json
{"status":"ready"}
```
