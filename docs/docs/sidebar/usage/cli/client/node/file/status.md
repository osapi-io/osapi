# Status

Check the deployment status of a file on the target host. Reports whether the
file is `in-sync`, `drifted`, or `missing`.

```bash
$ osapi client node file status --path /etc/app/app.conf

  Job ID:   550e8400-e29b-41d4-a716-446655440000
  Hostname: server1
  Path:     /etc/app/app.conf
  Status:   in-sync
  SHA256:   a1b2c3d4e5f6...
```

When a file has been modified on disk:

```bash
$ osapi client node file status --path /etc/app/app.conf

  Job ID:   550e8400-e29b-41d4-a716-446655440000
  Hostname: server1
  Path:     /etc/app/app.conf
  Status:   drifted
  SHA256:   9f8e7d6c5b4a...
```

When a file has not been deployed or was deleted:

```bash
$ osapi client node file status --path /etc/app/app.conf

  Job ID:   550e8400-e29b-41d4-a716-446655440000
  Hostname: server1
  Path:     /etc/app/app.conf
  Status:   missing
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node file status --path /etc/app/app.conf --json
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `--path`       | Filesystem path to check (**required**)                  |         |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
