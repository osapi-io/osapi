# Delete

Delete the DNS configuration that osapi manages for a network interface. The
agent removes the `/etc/netplan/osapi-dns.yaml` file and runs `netplan apply`:

```bash
$ osapi client node network dns delete \
    --interface eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true

  1 host: 1 changed
```

When you target all hosts:

```bash
$ osapi client node network dns delete \
    --interface eth0 \
    --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  server1   changed  true
  server2   skip

  2 hosts: 1 changed, 1 skipped

  Details:
  server2   unsupported platform
```

If osapi does not manage a DNS configuration for the interface, the command
returns `changed: false`.

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node network dns delete \
    --interface eth0 --json
{"results":[{"hostname":"web-01","changed":true,
"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--interface`  | Name of the network interface                            | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
