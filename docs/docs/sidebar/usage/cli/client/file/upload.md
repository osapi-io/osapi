# Upload

Upload a local file to the OSAPI Object Store for later deployment:

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

Re-uploading the same content is a no-op (`Changed: false`). If the file already
exists with different content, the upload is rejected with a 409 Conflict. Use
`--force` to overwrite:

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
