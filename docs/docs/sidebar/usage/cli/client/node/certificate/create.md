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

  HOSTNAME  NAME         CHANGED
  web-01    internal-ca  true
```

Broadcast to all hosts at once:

```bash
$ osapi client node certificate create --target _all \
    --name internal-ca --object internal-ca

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME         CHANGED
  web-01    internal-ca  true
  web-02    internal-ca  true
```

When some hosts are skipped (e.g., macOS agents), STATUS and ERROR columns are
added:

```bash
$ osapi client node certificate create --target _all \
    --name internal-ca --object internal-ca

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME         CHANGED  ERROR
  web-01    ok       internal-ca  true
  mac-01    skipped                        unsupported platform
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
