# Create

Create a new sysctl parameter on a target host. The value is written to
`/etc/sysctl.d/osapi-{key}.conf` and applied immediately via `sysctl -p`.
Idempotent: returns `changed: false` if already managed. Use `update` to change
an existing parameter:

```bash
$ osapi client node sysctl create --target web-01 \
    --key net.ipv4.ip_forward --value 1

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   KEY                   CHANGED
  web-01    changed  net.ipv4.ip_forward   true

  1 host: 1 changed
```

Broadcast to all hosts at once:

```bash
$ osapi client node sysctl create --target _all \
    --key vm.swappiness --value 10

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   KEY            CHANGED
  web-01    changed  vm.swappiness  true
  web-02    changed  vm.swappiness  true

  2 hosts: 2 changed
```

When some hosts are skipped (e.g., macOS agents):

```bash
$ osapi client node sysctl create --target _all \
    --key vm.swappiness --value 10

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   KEY            CHANGED
  web-01    changed  vm.swappiness  true
  mac-01    skip

  2 hosts: 1 changed, 1 skipped

  Details:
  mac-01    unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node sysctl create --target web-01 \
    --key vm.swappiness --value 10 --json
{"results":[{"hostname":"web-01","key":"vm.swappiness","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--key`        | Sysctl parameter key to create                           | required |
| `--value`      | Value to assign to the parameter                         | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
