# List

List all running processes on a target host:

```bash
$ osapi client node process list --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  PID   NAME      USER  STATE     CPU%  COMMAND
  web-01    ok      1     systemd   root  sleeping  0.0%  /sbin/init
  web-01    ok      245   sshd      root  sleeping  0.0%  sshd: /usr/sbin/sshd -D
  web-01    ok      1234  nginx     www   sleeping  2.3%  nginx: worker process

  1 host: 1 ok
```

When targeting all hosts:

```bash
$ osapi client node process list --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  PID   NAME      USER  STATE     CPU%  COMMAND
  web-01    ok      1     systemd   root  sleeping  0.0%  /sbin/init
  web-01    ok      1234  nginx     www   sleeping  2.3%  nginx: worker process
  web-02    ok      1     systemd   root  sleeping  0.0%  /sbin/init
  web-02    ok      5678  postgres  pg    sleeping  1.1%  postgres: writer process

  2 hosts: 2 ok
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node process list --target web-01 --json
{"results":[{"hostname":"web-01","processes":[{"pid":1,"name":"systemd","user":"root","state":"sleeping","cpu_percent":0.0,"mem_percent":0.1,"command":"/sbin/init"}],"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
