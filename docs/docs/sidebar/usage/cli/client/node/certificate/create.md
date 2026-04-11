# Create

Deploy a custom CA certificate to a target host. The PEM content must first be
uploaded to the Object Store. The certificate is written to
`/usr/local/share/ca-certificates/{name}.crt` and the system trust store is
rebuilt via `update-ca-certificates`. Fails if the name already exists -- use
`update` to replace:

```bash
$ osapi client node certificate create --target web-01 \
    --name internal-ca --object internal-ca

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME         CHANGED
  web-01    changed  internal-ca  true

  1 host: 1 changed
```

Broadcast to all hosts at once:

```bash
$ osapi client node certificate create --target _all \
    --name internal-ca --object internal-ca

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME         CHANGED
  web-01    changed  internal-ca  true
  web-02    changed  internal-ca  true

  2 hosts: 2 changed
```

When some hosts are skipped (e.g., macOS agents):

```bash
$ osapi client node certificate create --target _all \
    --name internal-ca --object internal-ca

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME         CHANGED
  web-01    changed  internal-ca  true
  mac-01    skip

  2 hosts: 1 changed, 1 skipped

  Details:
  mac-01    unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node certificate create --target web-01 \
    --name internal-ca --object internal-ca --json
{"results":[{"hostname":"web-01","name":"internal-ca","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Certificate name                                         | required |
| `--object`     | Object Store reference for the PEM file                  | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
