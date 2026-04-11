# Enable

Enable a systemd service to start on boot. Returns `changed: true` if the
service was not already enabled:

```bash
$ osapi client node service enable --target web-01 \
    --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME             CHANGED
  web-01    changed  nginx.service    true

  1 host: 1 changed
```

If the service is already enabled, `changed: false` is returned:

```bash
$ osapi client node service enable --target web-01 \
    --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  NAME             CHANGED
  web-01    ok      nginx.service    false

  1 host: 1 ok
```

Broadcast enable to all hosts:

```bash
$ osapi client node service enable --target _all \
    --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME             CHANGED
  web-01    changed  nginx.service    true
  web-02    changed  nginx.service    true

  2 hosts: 2 changed
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node service enable --target web-01 \
    --name nginx.service --json
{"results":[{"hostname":"web-01","name":"nginx.service","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Service name to enable                                   | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
