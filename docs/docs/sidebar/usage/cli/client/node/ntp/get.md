# Get

Get NTP sync status and configuration from a target host:

```bash
$ osapi client node ntp get --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  SYNCHRONIZED  SOURCE     SERVERS
  web-01    ok      yes           192.0.2.1  0.pool.ntp.org, 1.pool.ntp.org

  1 host: 1 ok
```

When targeting all hosts:

```bash
$ osapi client node ntp get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  SYNCHRONIZED  SOURCE     SERVERS
  web-01    ok      yes           192.0.2.1  0.pool.ntp.org, 1.pool.ntp.org
  web-02    ok      yes           192.0.2.1  0.pool.ntp.org, 1.pool.ntp.org

  2 hosts: 2 ok
```

When some hosts are skipped:

```bash
$ osapi client node ntp get --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  SYNCHRONIZED  SOURCE     SERVERS
  web-01    ok      yes           192.0.2.1  0.pool.ntp.org
  mac-01    skip

  2 hosts: 1 ok, 1 skipped

  Details:
  mac-01    unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node ntp get --target web-01 --json
{"results":[{"hostname":"web-01","synchronized":true,"stratum":2,
"offset":"+0.000123","current_source":"192.0.2.1",
"servers":["0.pool.ntp.org","1.pool.ntp.org"],"status":"ok"}],
"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
