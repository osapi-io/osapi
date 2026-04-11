# Update

Update NTP configuration on a target host. The server list is rewritten to
`/etc/chrony/conf.d/osapi-ntp.conf` and chrony is reloaded immediately. Fails if
NTP configuration is not currently managed -- use `create` first:

```bash
$ osapi client node ntp update --target web-01 \
    --servers ntp.example.com

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true

  1 host: 1 changed
```

If the server list is already identical, `changed: false` is returned and the
file is not rewritten:

```bash
$ osapi client node ntp update --target web-01 \
    --servers ntp.example.com

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED
  web-01    ok      false

  1 host: 1 ok
```

Broadcast to all hosts at once:

```bash
$ osapi client node ntp update --target _all \
    --servers 0.pool.ntp.org --servers 1.pool.ntp.org

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true
  web-02    changed  true

  2 hosts: 2 changed
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node ntp update --target web-01 \
    --servers ntp.example.com --json
{"results":[{"hostname":"web-01","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--servers`    | NTP server addresses (repeatable)                        | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
