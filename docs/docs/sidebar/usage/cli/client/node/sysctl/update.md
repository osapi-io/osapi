# Update

Update an existing sysctl parameter on a target host. The value is rewritten to
`/etc/sysctl.d/osapi-{key}.conf` and applied immediately via `sysctl -p`. Fails
if the key is not currently managed -- use `create` first:

```bash
$ osapi client node sysctl update --target web-01 \
    --key net.ipv4.ip_forward --value 0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   KEY                   CHANGED
  web-01    changed  net.ipv4.ip_forward   true

  1 host: 1 changed
```

If the parameter already has the requested value, `changed: false` is returned
and the file is not rewritten:

```bash
$ osapi client node sysctl update --target web-01 \
    --key net.ipv4.ip_forward --value 0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  KEY                   CHANGED
  web-01    ok      net.ipv4.ip_forward   false

  1 host: 1 ok
```

Broadcast to all hosts at once:

```bash
$ osapi client node sysctl update --target _all \
    --key vm.swappiness --value 20

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   KEY            CHANGED
  web-01    changed  vm.swappiness  true
  web-02    changed  vm.swappiness  true

  2 hosts: 2 changed
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node sysctl update --target web-01 \
    --key vm.swappiness --value 20 --json
{"results":[{"hostname":"web-01","key":"vm.swappiness","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--key`        | Sysctl parameter key to update                           | required |
| `--value`      | New value for the parameter                              | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
