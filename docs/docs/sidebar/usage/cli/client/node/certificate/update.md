# Update

Update an existing custom CA certificate on a target host with a new Object
Store reference. The PEM file is redeployed and `update-ca-certificates` is run
to rebuild the trust store. Fails if the certificate does not exist -- use
`create` first:

```bash
$ osapi client node certificate update --target web-01 \
    --name internal-ca --object internal-ca-v2

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME         CHANGED
  web-01    internal-ca  true
```

If the content has not changed (same SHA), `changed: false` is returned:

```bash
$ osapi client node certificate update --target web-01 \
    --name internal-ca --object internal-ca-v2

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME         CHANGED
  web-01    internal-ca  false
```

Broadcast to all hosts at once:

```bash
$ osapi client node certificate update --target _all \
    --name internal-ca --object internal-ca-v2

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME         CHANGED
  web-01    internal-ca  true
  web-02    internal-ca  true
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node certificate update --target web-01 \
    --name internal-ca --object internal-ca-v2 --json
{"results":[{"hostname":"web-01","name":"internal-ca","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Certificate name to update                               | required |
| `--object`     | New Object Store reference for the PEM file              | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
