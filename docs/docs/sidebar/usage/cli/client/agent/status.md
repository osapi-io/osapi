# Status

Show health status for all registered agents:

```bash
$ osapi client agent status

  TYPE    HOSTNAME   STATUS  CONDITIONS    AGE    CPU    MEM
  agent   web-01     Ready   DiskPressure  7h 6m  1.2%   96 MB
  agent   web-02     Ready   -             3h 2m  0.8%   82 MB
```

Use `--json` for the full health status JSON response:

```bash
$ osapi client agent status --json
```
