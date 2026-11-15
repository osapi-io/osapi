# Upload

Upload a file to the Object Store.

An upload of the same content changes nothing (`Changed: false`). If the file
exists with different content, the command rejects the upload with 409 Conflict.
Use `--force` to replace the file.

```bash
osapi client file upload
```

## Examples

```bash
osapi client file upload --name app.conf --file /tmp/app.conf
osapi client file upload --name app.conf.tmpl --file /tmp/app.conf.tmpl \
  --content-type template
osapi client file upload --name app.conf --file /tmp/app.conf --force
```

## Flags

| Flag             | Description                                                     | Default |
| ---------------- | --------------------------------------------------------------- | ------- |
| `--content-type` | File type: raw or template (default raw)                        | `raw`   |
| `--file`         | Path to the local file to upload (required)                     |         |
| `--force`        | Force upload even if file already exists with different content |         |
| `--name`         | Name for the file in the Object Store (required)                |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
