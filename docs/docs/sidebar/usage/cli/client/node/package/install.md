# Install

Install a package on the target host:

```bash
$ osapi client node package install --target web-01 --name nginx

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED  ERROR  NAME
  web-01    ok      true            nginx
```

Broadcast to all hosts at once:

```bash
$ osapi client node package install --target _all --name htop

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  ERROR                 NAME
  web-01    ok       true                            htop
  web-02    ok       true                            htop
  mac-01    skipped  false    unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node package install --target web-01 --name nginx --json
{"results":[{"hostname":"web-01","name":"nginx","changed":true,
"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Name of the package to install                           | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`   |
| `-j, --json`   | Output raw JSON response                                 |          |
