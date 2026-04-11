# Delete

Delete the OSAPI-managed DNS configuration for a network interface. Removes the
`/etc/netplan/osapi-dns.yaml` file and runs `netplan apply`:

```bash
$ osapi client node network dns delete \
    --interface-name eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true

  1 host: 1 changed
```

When targeting all hosts:

```bash
$ osapi client node network dns delete \
    --interface-name eth0 \
    --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  server1   changed  true
  server2   skip

  2 hosts: 1 changed, 1 skipped

  Details:
  server2   unsupported platform
```

Returns `changed: false` if no OSAPI-managed DNS configuration exists for the
interface.

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node network dns delete \
    --interface-name eth0 --json
{"results":[{"hostname":"web-01","changed":true,
"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag               | Description                                              | Default  |
| ------------------ | -------------------------------------------------------- | -------- |
| `--interface-name` | Name of the network interface                            | required |
| `-T, --target`     | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`       | Output raw JSON response                                 |          |
