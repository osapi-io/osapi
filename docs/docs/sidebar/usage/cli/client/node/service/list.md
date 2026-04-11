# List

List all systemd services on a target host:

```bash
$ osapi client node service list --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  NAME             ACTIVE  ENABLED
  web-01    ok      nginx.service    active  true
  web-01    ok      ssh.service      active  true
  web-01    ok      cron.service     active  true

  1 host: 1 ok
```

Target all hosts to list services across the fleet:

```bash
$ osapi client node service list --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  NAME             ACTIVE  ENABLED
  web-01    ok      nginx.service    active  true
  web-01    ok      ssh.service      active  true
  web-02    ok      nginx.service    active  true
  web-02    ok      ssh.service      active  true

  2 hosts: 2 ok
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node service list --target web-01 --json
{"results":[{"hostname":"web-01","status":"ok","services":[
{"name":"nginx.service","status":"active","enabled":true,
"description":"A high performance web server","pid":1234}
]}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`  |
| `-j, --json`   | Output raw JSON response                                 |         |
