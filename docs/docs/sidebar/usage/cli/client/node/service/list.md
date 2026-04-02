# List

List all systemd services on a target host:

```bash
$ osapi client node service list --target web-01

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME             STATUS    ENABLED  DESCRIPTION
  nginx.service    active    true     A high performance web server
  ssh.service      active    true     OpenBSD Secure Shell server
  cron.service     active    true     Regular background program processing
```

Target all hosts to list services across the fleet:

```bash
$ osapi client node service list --target _all

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  web-01
  NAME             STATUS    ENABLED  DESCRIPTION
  nginx.service    active    true     A high performance web server
  ssh.service      active    true     OpenBSD Secure Shell server

  web-02
  NAME             STATUS    ENABLED  DESCRIPTION
  nginx.service    active    true     A high performance web server
  ssh.service      active    true     OpenBSD Secure Shell server
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
