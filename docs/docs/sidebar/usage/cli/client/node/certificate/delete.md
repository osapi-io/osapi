# Delete

Delete a custom CA certificate from a target host. The certificate file is
removed from `/usr/local/share/ca-certificates/` and `update-ca-certificates` is
run to rebuild the trust store:

```bash
$ osapi client node certificate delete --target web-01 \
    --name internal-ca

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME         CHANGED
  internal-ca  true
```

If the certificate does not exist, `changed: false` is returned:

```bash
$ osapi client node certificate delete --target web-01 \
    --name internal-ca

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME         CHANGED
  internal-ca  false
```

Broadcast to all hosts:

```bash
$ osapi client node certificate delete --target _all \
    --name internal-ca

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME         CHANGED
  web-01    internal-ca  true
  web-02    internal-ca  true
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node certificate delete --target web-01 \
    --name internal-ca --json
{"results":[{"hostname":"web-01","name":"internal-ca","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Certificate name to delete                               | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
