# Signal

Send a signal to a specific process by PID:

```bash
$ osapi client node process signal --target web-01 \
    --pid 1234 --signal TERM

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  PID   SIGNAL  CHANGED
  web-01    1234  TERM    true
```

Broadcast a signal to a process on all hosts:

```bash
$ osapi client node process signal --target _all \
    --pid 1234 --signal HUP

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   PID   SIGNAL  CHANGED  ERROR
  web-01    ok       1234  HUP     true
  web-02    ok       1234  HUP     true
  mac-01    skipped                         unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node process signal --target web-01 \
    --pid 1234 --signal TERM --json
{"results":[{"hostname":"web-01","pid":1234,"signal":"TERM","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--pid`        | Process ID to signal (required)                          |         |
| `--signal`     | Signal name: TERM, KILL, HUP, INT, USR1, USR2, etc.      |         |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
