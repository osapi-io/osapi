# Updates

List packages that have newer versions available in the configured repositories:

```bash
$ osapi client node package updates --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME    CURRENT      NEW
  web-01    nginx   1.24.0-2     1.26.0-1
  web-01    curl    8.5.0-2      8.7.1-1
```

Broadcast to check for updates across all hosts:

```bash
$ osapi client node package updates --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME    CURRENT      NEW
  web-01    nginx   1.24.0-2     1.26.0-1
  web-01    curl    8.5.0-2      8.7.1-1
  web-02    nginx   1.24.0-2     1.26.0-1
```

When some hosts are skipped, STATUS and ERROR columns appear alongside data
columns:

```bash
$ osapi client node package updates --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   ERROR                 NAME    CURRENT      NEW
  web-01    ok                              nginx   1.24.0-2     1.26.0-1
  mac-01    skipped  unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node package updates --target web-01 --json
{"results":[{"hostname":"web-01","updates":[{"name":"nginx",
"current_version":"1.24.0-2","new_version":"1.26.0-1"}],
"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `-j, --json`   | Output raw JSON response                                 |         |
