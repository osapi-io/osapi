# Create

Make the NTP configuration on a target host. The agent writes the server list to
`/etc/chrony/conf.d/osapi-ntp.conf` and reloads chrony immediately. If osapi
already manages it, the command returns `changed: false`. To change the server
list, use `update`:

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

When some hosts are skipped, for example macOS agents:

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
