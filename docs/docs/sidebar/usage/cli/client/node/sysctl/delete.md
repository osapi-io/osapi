# Delete

Delete a sysctl parameter by removing its drop-in file from `/etc/sysctl.d/`.
The parameter value is **not** reset in the running kernel — it persists until
the next reboot or until you set a new value:

```bash
$ osapi client node sysctl delete --target web-01 \
    --key net.ipv4.ip_forward

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   KEY                   CHANGED
  web-01    changed  net.ipv4.ip_forward   true

  1 host: 1 changed
```

If the parameter does not exist, `changed: false` is returned:

```bash
$ osapi client node sysctl delete --target web-01 \
    --key net.ipv4.ip_forward

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  KEY                   CHANGED
  web-01    ok      net.ipv4.ip_forward   false

  1 host: 1 ok
```

Broadcast to all hosts:

```bash
$ osapi client node sysctl delete --target _all \
    --key vm.swappiness

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   KEY            CHANGED
  web-01    changed  vm.swappiness  true
  web-02    changed  vm.swappiness  true

  2 hosts: 2 changed
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node sysctl delete --target web-01 \
    --key net.ipv4.ip_forward --json
{"results":[{"hostname":"web-01","key":"net.ipv4.ip_forward","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--key`        | Sysctl parameter key to delete                           | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
