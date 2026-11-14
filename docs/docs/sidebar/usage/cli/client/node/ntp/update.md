# Update

Update the NTP configuration on a target host. The agent writes the server list
again to `/etc/chrony/conf.d/osapi-ntp.conf` and reloads chrony immediately. The
command fails if osapi does not manage the NTP configuration. Use `create`
first:

```bash
$ osapi client node ntp update --target web-01 \
    --servers ntp.example.com

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true

  1 host: 1 changed
```

If the server list is the same, the command returns `changed: false` and does
not write the file again:

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
