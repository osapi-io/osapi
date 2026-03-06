# Upload

Upload a local file to the OSAPI Object Store for later deployment:

```bash
$ osapi client file upload --name app.conf --file /tmp/app.conf

  Name:   app.conf
  SHA256: a1b2c3d4e5f6...
  Size:   1234
```

Upload a template file:

```bash
$ osapi client file upload --name app.conf.tmpl --file /tmp/app.conf.tmpl
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client file upload --name app.conf --file /tmp/app.conf --json
```

## Flags

| Flag         | Description                                          | Default |
| ------------ | ---------------------------------------------------- | ------- |
| `--name`     | Name for the file in the Object Store (**required**) |         |
| `--file`     | Path to the local file to upload (**required**)      |         |
| `-j, --json` | Output raw JSON response                             |         |
