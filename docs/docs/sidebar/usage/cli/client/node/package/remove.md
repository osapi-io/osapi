# Remove

Remove a package from the target host:

```bash
$ osapi client node package remove --target web-01 --name nginx

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME    CHANGED
  nginx   true
```

Broadcast to all hosts at once:

```bash
$ osapi client node package remove --target _all --name nginx

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME    CHANGED
  web-01    nginx   true
  web-02    nginx   true
```

When some hosts are skipped (e.g., macOS agents), STATUS and ERROR columns are
added:

```bash
$ osapi client node package remove --target _all --name nginx

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME    CHANGED  ERROR
  web-01    ok       nginx   true
  mac-01    skipped                    unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node package remove --target web-01 --name nginx --json
{"results":[{"hostname":"web-01","name":"nginx","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Name of the package to remove                            | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`   |
| `-j, --json`   | Output raw JSON response                                 |          |
