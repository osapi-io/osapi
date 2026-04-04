# Delete

Delete the OSAPI-managed DNS configuration for a network interface. Removes the
`/etc/netplan/osapi-dns.yaml` file and runs `netplan apply`:

```bash
$ osapi client node network dns delete \
    --interface-name eth0

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  STATUS  CHANGED  ERROR
  ok      true
```

When targeting all hosts, HOSTNAME is shown. STATUS and ERROR columns appear
when any host has an error or is skipped:

```bash
$ osapi client node network dns delete \
    --interface-name eth0 \
    --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  ERROR
  server1   ok       true
  server2   skipped           unsupported platform
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
