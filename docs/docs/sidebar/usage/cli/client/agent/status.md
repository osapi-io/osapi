# Status

Show health status for all registered agents:

```bash
$ osapi client agent status

  TYPE    HOSTNAME   STATUS  CONDITIONS    AGE    CPU    MEM
  agent   web-01     Ready   DiskPressure  7h 6m  1.2%   96 MB
          ├─ heartbeat               ok
          └─ metrics                 ok http://0.0.0.0:9091
  agent   web-02     Ready   -             3h 2m  0.8%   82 MB
          ├─ heartbeat               ok
          └─ metrics                 ok http://0.0.0.0:9091
```

Use `--json` for the full health status JSON response:

```bash
$ osapi client agent status --json
```
