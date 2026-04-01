# Get

Get details for a specific package by name:

```bash
$ osapi client node package get --target web-01 --name nginx

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME    VERSION    STATUS      SIZE
  nginx   1.24.0-2   installed   1.2 MB
```

Broadcast to see the package across all hosts:

```bash
$ osapi client node package get --target _all --name nginx

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME    VERSION    STATUS      SIZE
  web-01    nginx   1.24.0-2   installed   1.2 MB
  web-02    nginx   1.24.0-2   installed   1.2 MB
```

When some hosts are skipped, STATUS and ERROR columns appear alongside data
columns:

```bash
$ osapi client node package get --target _all --name nginx

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   ERROR                 NAME    VERSION    STATUS      SIZE
  web-01    ok                              nginx   1.24.0-2   installed   1.2 MB
  mac-01    skipped  unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node package get --target web-01 --name nginx --json
{"results":[{"hostname":"web-01","packages":[{"name":"nginx","version":
"1.24.0-2","status":"installed","size":1258291}],"status":"ok"}],
"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Name of the package                                      | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`   |
| `-j, --json`   | Output raw JSON response                                 |          |
