# Create

Create NTP configuration on a target host. The server list is written to
`/etc/chrony/conf.d/osapi-ntp.conf` and chrony is reloaded immediately.
Idempotent: returns `changed: false` if already managed. Use `update` to change
the server list:

```bash
$ osapi client node ntp create --target web-01 \
    --servers 0.pool.ntp.org --servers 1.pool.ntp.org

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true

  1 host: 1 changed
```

Broadcast to all hosts at once:

```bash
$ osapi client node ntp create --target _all \
    --servers 0.pool.ntp.org --servers 1.pool.ntp.org

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true
  web-02    changed  true

  2 hosts: 2 changed
```

When some hosts are skipped (e.g., macOS agents):

```bash
$ osapi client node ntp create --target _all \
    --servers 0.pool.ntp.org

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true
  mac-01    skip

  2 hosts: 1 changed, 1 skipped

  Details:
  mac-01    unsupported platform
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node ntp create --target web-01 \
    --servers 0.pool.ntp.org --json
{"results":[{"hostname":"web-01","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--servers`    | NTP server addresses (repeatable)                        | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
