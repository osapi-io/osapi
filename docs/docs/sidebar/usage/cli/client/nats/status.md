# Status

Show health status for all registered NATS servers:

```bash
$ osapi client nats status

  TYPE  HOSTNAME        STATUS  CONDITIONS  AGE    CPU    MEM
  nats  nats-server-01  Ready   -           7h 6m  0.3%   64 MB
        ├─ heartbeat               ok
        └─ metrics                 ok
```

Use `--json` for the full health status JSON response:

```bash
$ osapi client nats status --json
```
