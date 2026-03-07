# Delete

Delete a file from the Object Store:

```bash
$ osapi client file delete --name app.conf

  Name:    app.conf
  Deleted: true
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client file delete --name app.conf --json
```

## Flags

| Flag         | Description                                         | Default |
| ------------ | --------------------------------------------------- | ------- |
| `--name`     | Name of the file in the Object Store (**required**) |         |
| `-j, --json` | Output raw JSON response                            |         |
