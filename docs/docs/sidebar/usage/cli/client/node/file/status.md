# Status

Check the deployment status of a file on the target host. Reports whether the
file is `in-sync`, `drifted`, or `missing`.

```bash
$ osapi client node file status --path /etc/app/app.conf

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  PATH               FILE STATUS  SHA256
  server1   ok      /etc/app/app.conf  in-sync      a1b2c3d4e5f6...

  1 host: 1 ok
```

When a file has been modified on disk:

```bash
$ osapi client node file status --path /etc/app/app.conf

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  PATH               FILE STATUS  SHA256
  server1   ok      /etc/app/app.conf  drifted      9f8e7d6c5b4a...

  1 host: 1 ok
```

When a file has not been deployed or was deleted:

```bash
$ osapi client node file status --path /etc/app/app.conf

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  PATH               FILE STATUS
  server1   ok      /etc/app/app.conf  missing

  1 host: 1 ok
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
