# Query

Query journal log entries on a target host:

```bash
$ osapi client node log query --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  TIMESTAMP                  PRIORITY  UNIT           MESSAGE
  2026-01-01T00:00:01+00:00  info      sshd.service   Accepted publickey for ...
  2026-01-01T00:00:02+00:00  notice    kernel         Linux version 6.1.0 ...
  2026-01-01T00:00:03+00:00  err       nginx.service  connect() failed (111...
```

Filter by priority and time window:

```bash
$ osapi client node log query --target web-01 \
  --lines 50 --since 1h --priority err

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  TIMESTAMP                  PRIORITY  UNIT           MESSAGE
  2026-01-01T00:59:01+00:00  err       nginx.service  connect() failed (111...
```

When targeting all hosts:

```bash
$ osapi client node log query --target _all --lines 5

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  web-01
  TIMESTAMP                  PRIORITY  UNIT           MESSAGE
  2026-01-01T00:00:01+00:00  info      sshd.service   Accepted publickey for ...

  web-02
  TIMESTAMP                  PRIORITY  UNIT           MESSAGE
  2026-01-01T00:00:02+00:00  notice    kernel         Linux version 6.1.0 ...
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node log query --target web-01 --lines 1 --json
{"results":[{"hostname":"web-01","status":"ok","entries":[{"timestamp":
"2026-01-01T00:00:01+00:00","unit":"sshd.service","priority":"info",
"message":"Accepted publickey for user from 1.2.3.4 port 22 ssh2",
"pid":1234,"hostname":"web-01"}]}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `--lines`      | Maximum number of log lines to return                    | `100`   |
| `--since`      | Return entries since this time (e.g., `1h`,              |         |
|                | `2026-01-01 00:00:00`)                                   |         |
| `--priority`   | Filter by priority level (e.g., `err`, `warning`,        |         |
|                | `info`)                                                  |         |
| `-j, --json`   | Output raw JSON response                                 |         |
