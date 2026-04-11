# Delete

Remove an OSAPI-managed Netplan interface configuration from a target host. Only
files with the `osapi-` prefix are managed and can be deleted:

```bash
$ osapi client node network interface delete \
    --target web-01 --name eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME    CHANGED
  web-01    changed  eth0    true

  1 host: 1 changed
```

Broadcast delete to all hosts:

```bash
$ osapi client node network interface delete \
    --target _all --name eth1

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME    CHANGED
  web-01    changed  eth1    true
  web-02    changed  eth1    true

  2 hosts: 2 changed
```

When some hosts are skipped:

```bash
$ osapi client node network interface delete \
    --target _all --name eth1

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   NAME    CHANGED
  web-01    changed  eth1    true
  mac-01    skip

  2 hosts: 1 changed, 1 skipped

  Details:
  mac-01    unsupported platform
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
