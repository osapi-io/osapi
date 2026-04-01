# List

List all installed packages on a target host:

```bash
$ osapi client node package list --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME      VERSION      STATUS      SIZE
  bash      5.2.21-2     installed   7.4 MB
  nginx     1.24.0-2     installed   1.2 MB
  curl      8.5.0-2      installed   512.0 KB
```

Target all hosts to list packages across the fleet:

```bash
$ osapi client node package list --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  web-01
  NAME      VERSION      STATUS      SIZE
  bash      5.2.21-2     installed   7.4 MB
  nginx     1.24.0-2     installed   1.2 MB

  web-02
  NAME      VERSION      STATUS      SIZE
  bash      5.2.21-2     installed   7.4 MB
```

Target by label to list packages on a group of servers:

```bash
$ osapi client node package list --target group:web
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node package list --target web-01 --json
{"results":[{"hostname":"web-01","packages":[{"name":"bash","version":"5.2.21-2","status":"installed","size":7405568}],"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `-j, --json`   | Output raw JSON response                                 |         |
