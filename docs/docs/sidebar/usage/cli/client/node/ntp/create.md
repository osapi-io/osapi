# Create

Create NTP configuration on a target host. The server list is written to
`/etc/chrony/conf.d/osapi-ntp.conf` and chrony is reloaded immediately. Fails
if NTP configuration is already managed -- use `update` to change the server
list:

```bash
$ osapi client node ntp create --target web-01 \
    --servers 0.pool.ntp.org --servers 1.pool.ntp.org

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED  ERROR
  web-01    ok      true
```

Broadcast to all hosts at once:

```bash
$ osapi client node ntp create --target _all \
    --servers 0.pool.ntp.org --servers 1.pool.ntp.org

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED  ERROR
  web-01    ok      true
  web-02    ok      true
```

When some hosts are skipped (e.g., macOS agents):

```bash
$ osapi client node ntp create --target _all \
    --servers 0.pool.ntp.org

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  ERROR
  web-01    ok       true
  mac-01    skipped           unsupported platform
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
