# Upload

Upload a local file to the osapi Object Store, to deploy it later:

```bash
$ osapi client file upload --name app.conf --file /tmp/app.conf

  Name:         app.conf
  SHA256:       a1b2c3d4e5f6...
  Size:         1234
  Changed:      true
  Content-Type: raw
```

Upload a template file:

```bash
$ osapi client file upload --name app.conf.tmpl --file /tmp/app.conf.tmpl \
    --content-type template
```

An upload of the same content changes nothing (`Changed: false`). If the file
exists with different content, the command rejects the upload with 409 Conflict.
Use `--force` to replace the file:

```bash
$ osapi client file upload --name app.conf --file /tmp/app.conf --force
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client file upload --name app.conf --file /tmp/app.conf --json
```

## Flags

| Flag             | Description                                          | Default |
| ---------------- | ---------------------------------------------------- | ------- |
| `--name`         | Name for the file in the Object Store (**required**) |         |
| `--file`         | Path to the local file to upload (**required**)      |         |
| `--content-type` | File type: `raw` or `template`                       | `raw`   |
| `--force`        | Force upload even if file exists with different data |         |
| `-j, --json`     | Output raw JSON response                             |         |
