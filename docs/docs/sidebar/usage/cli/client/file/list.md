# List

List all files stored in the OSAPI Object Store:

```bash
$ osapi client file list

  Files (3)
  NAME             SHA256            SIZE
  app.conf         a1b2c3d4e5f6...   1234
  app.conf.tmpl    f6e5d4c3b2a1...   567
  nginx.conf       1a2b3c4d5e6f...   2048
```

When no files are stored:

```bash
$ osapi client file list

  No files found.
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client file list --json
```

## Flags

| Flag         | Description              | Default |
| ------------ | ------------------------ | ------- |
| `-j, --json` | Output raw JSON response |         |
