# Shutdown

Schedule a shutdown on the target host. The agent calls `shutdown -h` with the
configured delay. When `--delay` is 0, the shutdown happens immediately:

```bash
$ osapi client node power shutdown --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  STATUS  CHANGED  ERROR  ACTION    DELAY
  ok      true            shutdown  0
```

Shutdown with a 60-second delay and a broadcast message:

```bash
$ osapi client node power shutdown --target web-01 \
    --delay 60 --message "Scheduled maintenance shutdown"

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  STATUS  CHANGED  ERROR  ACTION    DELAY
  ok      true            shutdown  60
```

Broadcast shutdown to all hosts at once:

```bash
$ osapi client node power shutdown --target _all --delay 30

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  ERROR                ACTION    DELAY
  web-01    ok       true                           shutdown  30
  mac-01    skipped           unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node power shutdown --target web-01 --json
{"results":[{"hostname":"web-01","action":"shutdown","delay":0,"changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--delay`      | Seconds to wait before shutting down                     | `0`     |
| `--message`    | Optional message to broadcast before shutdown            |         |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `-j, --json`   | Output raw JSON response                                 |         |
