# Delete

Remove an OSAPI-managed Netplan interface configuration from a target host. Only
files with the `osapi-` prefix are managed and can be deleted:

```bash
$ osapi client node network interface delete \
    --target web-01 --name eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME    CHANGED
  eth0    true
```

Broadcast delete to all hosts:

```bash
$ osapi client node network interface delete \
    --target _all --name eth1

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  NAME    CHANGED
  web-01    eth1    true
  web-02    eth1    true
```

When some hosts are skipped, STATUS and ERROR columns are added:

```bash
$ osapi client node network interface delete \
    --target _all --name eth1

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME    CHANGED  ERROR
  web-01    ok       eth1    true
  mac-01    skipped                   unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node network interface delete \
    --target web-01 --name eth0 --json
{"results":[{"hostname":"web-01","name":"eth0",
"changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Interface name to delete                                 | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
