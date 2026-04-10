# List

List all running processes on a target host:

```bash
$ osapi client node process list --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  PID   NAME      USER  STATE     CPU%  MEM%  COMMAND
  web-01    1     systemd   root  sleeping  0.0%  0.1%  /sbin/init
  web-01    245   sshd      root  sleeping  0.0%  0.2%  sshd: /usr/sbin/sshd -D
  web-01    1234  nginx     www   sleeping  2.3%  1.5%  nginx: worker process
```

When targeting all hosts:

```bash
$ osapi client node process list --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  PID   NAME      USER  STATE     CPU%  MEM%  COMMAND
  web-01    1     systemd   root  sleeping  0.0%  0.1%  /sbin/init
  web-01    1234  nginx     www   sleeping  2.3%  1.5%  nginx: worker process
  web-02    1     systemd   root  sleeping  0.0%  0.1%  /sbin/init
  web-02    5678  postgres  pg    sleeping  1.1%  3.2%  postgres: writer process
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
