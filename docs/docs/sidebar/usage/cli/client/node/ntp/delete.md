# Delete

Delete the NTP configuration: the agent removes the drop-in file from
`/etc/chrony/conf.d/`, then reloads chrony. The system then uses the other
chrony configuration that stays:

```bash
$ osapi client node ntp delete --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true

  1 host: 1 changed
```

If the configuration does not exist, the command returns `changed: false`:

```bash
$ osapi client node ntp delete --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED
  web-01    ok      false

  1 host: 1 ok
```

Broadcast to all hosts:

```bash
$ osapi client node ntp delete --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true
  web-02    changed  true

  2 hosts: 2 changed
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node ntp delete --target web-01 --json
{"results":[{"hostname":"web-01","changed":true,"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
