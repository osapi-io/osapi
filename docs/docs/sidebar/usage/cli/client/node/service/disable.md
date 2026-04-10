# Disable

Disable a systemd service from starting on boot. Returns `changed: true` if the
service was previously enabled:

```bash
$ osapi client node service disable --target web-01 \
    --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME             CHANGED
  web-01    nginx.service    true
```

If the service is already disabled, `changed: false` is returned:

```bash
$ osapi client node service disable --target web-01 \
    --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME             CHANGED
  web-01    nginx.service    false
```

Broadcast disable to all hosts:

```bash
$ osapi client node service disable --target _all \
    --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME             CHANGED
  web-01    nginx.service    true
  web-02    nginx.service    true
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node service disable --target web-01 \
    --name nginx.service --json
{"results":[{"hostname":"web-01","name":"nginx.service","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Service name to disable                                  | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
