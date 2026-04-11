# Remove

Remove a package from the target host:

```bash
$ osapi client node package remove --target web-01 --name nginx

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  NAME
  web-01    changed  true     nginx

  1 host: 1 changed
```

Broadcast to all hosts at once:

```bash
$ osapi client node package remove --target _all --name nginx

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  NAME
  web-01    changed  true     nginx
  web-02    changed  true     nginx
  mac-01    skip

  3 hosts: 2 changed, 1 skipped

  Details:
  mac-01    unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node package remove --target web-01 --name nginx --json
{"results":[{"hostname":"web-01","name":"nginx","changed":true,
"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Name of the package to remove                            | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`   |
| `-j, --json`   | Output raw JSON response                                 |          |
