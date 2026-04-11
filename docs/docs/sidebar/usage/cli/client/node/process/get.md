# Get

Get detailed information about a specific process by PID:

```bash
$ osapi client node process get --target web-01 --pid 1234

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  PID   NAME   USER  STATE     CPU%  COMMAND
  web-01    ok      1234  nginx  www   sleeping  2.3%  nginx: worker process

  1 host: 1 ok
```

When targeting all hosts:

```bash
$ osapi client node process get --target _all --pid 1

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  PID  NAME     USER  STATE     CPU%  COMMAND
  web-01    ok      1    systemd  root  sleeping  0.0%  /sbin/init
  web-02    ok      1    systemd  root  sleeping  0.0%  /sbin/init

  2 hosts: 2 ok
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node process get --target web-01 --pid 1234 --json
{"results":[{"hostname":"web-01","process":{"pid":1234,"name":"nginx","user":"www","state":"sleeping","cpu_percent":2.3,"mem_percent":1.5,"command":"nginx: worker process"},"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--pid`        | Process ID to inspect (required)                         |         |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
