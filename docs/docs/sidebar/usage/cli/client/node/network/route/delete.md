# Delete

Remove the OSAPI-managed static routes for an interface on a target host. Only
OSAPI-managed files (with the `osapi-` prefix) can be deleted:

```bash
$ osapi client node network route delete \
    --target web-01 --interface eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  INTERFACE  CHANGED
  eth0       true
```

Broadcast delete to all hosts:

```bash
$ osapi client node network route delete \
    --target _all --interface eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  INTERFACE  CHANGED
  web-01    eth0       true
  web-02    eth0       true
```

When some hosts are skipped, STATUS and ERROR columns are added:

```bash
$ osapi client node network route delete \
    --target _all --interface eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   INTERFACE  CHANGED  ERROR
  web-01    ok       eth0       true
  mac-01    skipped                      unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node network route delete \
    --target web-01 --interface eth0 --json
{"results":[{"hostname":"web-01","interface":"eth0",
"changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--interface`  | Interface name                                           | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
