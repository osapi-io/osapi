# Update

Update an existing service unit file on a target host with a new Object Store
reference. The unit file is redeployed and `systemctl daemon-reload` is run.
Fails if the unit file does not exist -- use `create` first:

```bash
$ osapi client node service update --target web-01 \
    --name myapp.service --object myapp-unit-v2

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME             CHANGED
  web-01    myapp.service    true
```

If the content has not changed (same SHA), `changed: false` is returned:

```bash
$ osapi client node service update --target web-01 \
    --name myapp.service --object myapp-unit-v2

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME             CHANGED
  web-01    myapp.service    false
```

Broadcast to all hosts at once:

```bash
$ osapi client node service update --target _all \
    --name myapp.service --object myapp-unit-v2

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME             CHANGED
  web-01    myapp.service    true
  web-02    myapp.service    true
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node service update --target web-01 \
    --name myapp.service --object myapp-unit-v2 --json
{"results":[{"hostname":"web-01","name":"myapp.service","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Service unit name to update                              | required |
| `--object`     | New Object Store reference for the unit file             | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
