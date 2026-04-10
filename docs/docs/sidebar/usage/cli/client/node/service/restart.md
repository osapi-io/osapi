# Restart

Restart a systemd service on a target host. Always returns `changed: true`
because the service is restarted regardless of its current state:

```bash
$ osapi client node service restart --target web-01 \
    --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME             CHANGED
  web-01    nginx.service    true
```

Broadcast restart to all hosts:

```bash
$ osapi client node service restart --target _all \
    --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME             CHANGED
  web-01    nginx.service    true
  web-02    nginx.service    true
```

When some hosts are skipped (e.g., macOS agents), STATUS and ERROR columns are
added:

```bash
$ osapi client node service restart --target _all \
    --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME             CHANGED  ERROR
  web-01    ok       nginx.service    true
  mac-01    skipped                            unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node service restart --target web-01 \
    --name nginx.service --json
{"results":[{"hostname":"web-01","name":"nginx.service","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Service name to restart                                  | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
