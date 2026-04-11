# Unit

Query journal log entries for a specific systemd unit on a target host:

```bash
$ osapi client node log unit --target web-01 --name sshd.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  TIMESTAMP                  UNIT          MESSAGE
  web-01    ok      2026-01-01T00:00:01+00:00  sshd.service  Accepted publickey for ...
  web-01    ok      2026-01-01T00:00:02+00:00  sshd.service  pam_unix(sshd:session): ...
  web-01    ok      2026-01-01T00:00:03+00:00  sshd.service  Disconnected from user ...

  1 host: 1 ok
```

Filter to recent errors only:

```bash
$ osapi client node log unit --target web-01 --name nginx.service \
  --lines 20 --since 30m --priority err

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  TIMESTAMP                  UNIT           MESSAGE
  web-01    ok      2026-01-01T00:31:15+00:00  nginx.service  connect() failed (111...

  1 host: 1 ok
```

When targeting all hosts:

```bash
$ osapi client node log unit --target _all --name sshd.service --lines 5

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  TIMESTAMP                  UNIT          MESSAGE
  web-01    ok      2026-01-01T00:00:01+00:00  sshd.service  Accepted publickey for ...
  web-02    ok      2026-01-01T00:00:02+00:00  sshd.service  Accepted publickey for ...

  2 hosts: 2 ok
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node log unit --target web-01 --name sshd.service \
  --lines 1 --json
{"results":[{"hostname":"web-01","status":"ok","entries":[{"timestamp":
"2026-01-01T00:00:01+00:00","unit":"sshd.service","priority":"info",
"message":"Accepted publickey for user from 1.2.3.4 port 22 ssh2",
"pid":1234,"hostname":"web-01"}]}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `--name`       | Systemd unit name (required, e.g., `sshd.service`)       |         |
| `--lines`      | Maximum number of log lines to return                    | `100`   |
| `--since`      | Return entries since this time (e.g., `1h`,              |         |
|                | `2026-01-01 00:00:00`)                                   |         |
| `--priority`   | Filter by priority level (e.g., `err`, `warning`,        |         |
|                | `info`)                                                  |         |
| `-j, --json`   | Output raw JSON response                                 |         |
