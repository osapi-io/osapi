# Restart

Restart a systemd service on a target host. Always returns `changed: true`
because the service is restarted regardless of its current state:

```bash
$ osapi client node service restart --target web-01 \
    --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME             CHANGED
  web-01    changed  nginx.service    true

  1 host: 1 changed
```

Broadcast restart to all hosts:

```bash
$ osapi client node service restart --target _all \
    --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME             CHANGED
  web-01    changed  nginx.service    true
  web-02    changed  nginx.service    true

  2 hosts: 2 changed
```

When some hosts are skipped (e.g., macOS agents):

```bash
$ osapi client node service restart --target _all \
    --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME             CHANGED
  web-01    changed  nginx.service    true
  mac-01    skip

  2 hosts: 1 changed, 1 skipped

  Details:
  mac-01    unsupported platform
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
