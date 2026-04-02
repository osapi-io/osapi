# Create

Deploy a new service unit file to a target host. The unit file content must
first be uploaded to the Object Store. The file is written to
`/etc/systemd/system/{name}` and `systemctl daemon-reload` is run. Fails if the
name already exists -- use `update` to replace:

```bash
$ osapi client node service create --target web-01 \
    --name myapp.service --object myapp-unit

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME             CHANGED
  myapp.service    true
```

Broadcast to all hosts at once:

```bash
$ osapi client node service create --target _all \
    --name myapp.service --object myapp-unit

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME             CHANGED
  web-01    myapp.service    true
  web-02    myapp.service    true
```

When some hosts are skipped (e.g., macOS agents), STATUS and ERROR columns are
added:

```bash
$ osapi client node service create --target _all \
    --name myapp.service --object myapp-unit

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME             CHANGED  ERROR
  web-01    ok       myapp.service    true
  mac-01    skipped                            unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node service create --target web-01 \
    --name myapp.service --object myapp-unit --json
{"results":[{"hostname":"web-01","name":"myapp.service","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Service unit name                                        | required |
| `--object`     | Object Store reference for the unit file                 | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
