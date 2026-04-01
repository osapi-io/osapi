# Delete

Delete NTP configuration by removing the drop-in file from
`/etc/chrony/conf.d/`. Chrony is reloaded after removal. The system falls back
to whatever other chrony configuration remains:

```bash
$ osapi client node ntp delete --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  STATUS  CHANGED  ERROR
  ok      true
```

If the configuration does not exist, `changed: false` is returned:

```bash
$ osapi client node ntp delete --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  STATUS  CHANGED  ERROR
  ok      false
```

Broadcast to all hosts:

```bash
$ osapi client node ntp delete --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  ERROR
  web-01    ok       true
  web-02    ok       true
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
