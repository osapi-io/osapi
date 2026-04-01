# Source

List available log sources (syslog identifiers) on a target host:

```bash
$ osapi client node log source --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  SOURCE
  cron
  kernel
  nginx
  sshd
  systemd
```

When targeting all hosts:

```bash
$ osapi client node log source --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  web-01
  SOURCE
  cron
  kernel
  nginx
  sshd

  web-02
  SOURCE
  kernel
  systemd
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node log source --target web-01 --json
{"results":[{"hostname":"web-01","status":"ok","sources":["cron",
"kernel","nginx","sshd","systemd"]}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
