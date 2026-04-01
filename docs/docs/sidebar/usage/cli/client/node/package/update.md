# Update

Refresh the package source lists on the target host. This is equivalent to
running `apt-get update` -- it does not upgrade any packages:

```bash
$ osapi client node package update --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  STATUS  CHANGED  ERROR
  ok      true
```

Broadcast to refresh sources across all hosts:

```bash
$ osapi client node package update --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  ERROR
  web-01    ok       true
  web-02    ok       true
  mac-01    skipped  false    unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node package update --target web-01 --json
{"results":[{"hostname":"web-01","changed":true,"status":"ok"}],
"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `-j, --json`   | Output raw JSON response                                 |         |
