# Create

Deploy a new service unit file to a target host. First, upload the content of
the unit file to the Object Store. The agent writes the file to
`/etc/systemd/system/{name}` and runs `systemctl daemon-reload`. The command
fails if the name exists. To replace it, use `update`:

```bash
$ osapi client node service create --target web-01 \
    --name myapp.service --object myapp-unit

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME             CHANGED
  web-01    changed  myapp.service    true

  1 host: 1 changed
```

Broadcast to all hosts at once:

```bash
$ osapi client node service create --target _all \
    --name myapp.service --object myapp-unit

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME             CHANGED
  web-01    changed  myapp.service    true
  web-02    changed  myapp.service    true

  2 hosts: 2 changed
```

When some hosts are skipped, for example macOS agents:

```bash
$ osapi client node service create --target _all \
    --name myapp.service --object myapp-unit

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME             CHANGED
  web-01    changed  myapp.service    true
  mac-01    skip

  2 hosts: 1 changed, 1 skipped

  Details:
  mac-01    unsupported platform
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
