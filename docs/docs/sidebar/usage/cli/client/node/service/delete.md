# Delete

Delete a service unit file from a target host. The service is stopped and
disabled if running, the unit file is removed from `/etc/systemd/system/`, and
`systemctl daemon-reload` is run:

```bash
$ osapi client node service delete --target web-01 \
    --name myapp.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME             CHANGED
  myapp.service    true
```

If the unit file does not exist, `changed: false` is returned:

```bash
$ osapi client node service delete --target web-01 \
    --name myapp.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME             CHANGED
  myapp.service    false
```

Broadcast to all hosts:

```bash
$ osapi client node service delete --target _all \
    --name myapp.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME             CHANGED
  web-01    myapp.service    true
  web-02    myapp.service    true
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node service delete --target web-01 \
    --name myapp.service --json
{"results":[{"hostname":"web-01","name":"myapp.service","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Service unit name to delete                              | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
