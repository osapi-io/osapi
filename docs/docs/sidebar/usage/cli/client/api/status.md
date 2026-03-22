# Status

Show health status for the controller:

```bash
$ osapi client controller status

  TYPE        HOSTNAME       STATUS  CONDITIONS  AGE    CPU    MEM
  controller  controller-01  Ready   -           7h 6m  2.1%   128 MB
              ├─ api                     ok http://0.0.0.0:8080
              ├─ heartbeat               ok
              ├─ kv (connectivity)       ok
              ├─ metrics                 ok http://0.0.0.0:9090
              ├─ nats (connectivity)     ok
              ├─ notifier                ok
              └─ tracing                 ok
```

Use `--json` for the full health status JSON response:

```bash
$ osapi client controller status --json
```
