# List

List all installed packages on a target host:

```bash
$ osapi client node package list --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  NAME      VERSION      PKG STATUS
  web-01    ok      bash      5.2.21-2     installed
  web-01    ok      nginx     1.24.0-2     installed
  web-01    ok      curl      8.5.0-2      installed

  1 host: 1 ok
```

Target all hosts to list packages across the fleet:

```bash
$ osapi client node package list --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  NAME      VERSION      PKG STATUS
  web-01    ok      bash      5.2.21-2     installed
  web-01    ok      nginx     1.24.0-2     installed
  web-02    ok      bash      5.2.21-2     installed

  2 hosts: 2 ok
```

When some hosts are skipped:

```bash
$ osapi client node package list --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  NAME      VERSION      PKG STATUS
  web-01    ok      bash      5.2.21-2     installed
  web-01    ok      nginx     1.24.0-2     installed
  mac-01    skip

  2 hosts: 1 ok, 1 skipped

  Details:
  mac-01    unsupported platform
```

Target by label to list packages on a group of servers:

```bash
$ osapi client node package list --target group:web
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node package list --target web-01 --json
{"results":[{"hostname":"web-01","packages":[{"name":"bash","version":
"5.2.21-2","status":"installed","size":7405568}],"status":"ok"}],
"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `-j, --json`   | Output raw JSON response                                 |         |

## Updates available

List packages that have newer versions available in the configured repositories:

```bash
$ osapi client node package list --updates --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  NAME    CURRENT      NEW
  web-01    ok      nginx   1.24.0-2     1.26.0-1
  web-01    ok      curl    8.5.0-2      8.7.1-1

  1 host: 1 ok
```

Broadcast to check for updates across all hosts:

```bash
$ osapi client node package list --updates --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  NAME    CURRENT      NEW
  web-01    ok      nginx   1.24.0-2     1.26.0-1
  web-01    ok      curl    8.5.0-2      8.7.1-1
  web-02    ok      nginx   1.24.0-2     1.26.0-1

  2 hosts: 2 ok
```

When some hosts are skipped:

```bash
$ osapi client node package list --updates --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  NAME    CURRENT      NEW
  web-01    ok      nginx   1.24.0-2     1.26.0-1
  mac-01    skip

  2 hosts: 1 ok, 1 skipped

  Details:
  mac-01    unsupported platform
```

### JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node package list --updates --target web-01 --json
{"results":[{"hostname":"web-01","updates":[{"name":"nginx",
"current_version":"1.24.0-2","new_version":"1.26.0-1"}],
"status":"ok"}],"job_id":"..."}
```

### Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `-j, --json`   | Output raw JSON response                                 |         |
