# Get

Get metadata for a specific file in the Object Store:

```bash
$ osapi client file get --name app.conf

  Name:   app.conf
  SHA256: a1b2c3d4e5f6...
  Size:   1234
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client file get --name app.conf --json
```

## Flags

| Flag         | Description                                         | Default |
| ------------ | --------------------------------------------------- | ------- |
| `--name`     | Name of the file in the Object Store (**required**) |         |
| `-j, --json` | Output raw JSON response                            |         |
