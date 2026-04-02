# Get

Get details for a specific systemd service on a target host:

```bash
$ osapi client node service get --target web-01 --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME             STATUS  ENABLED  DESCRIPTION                       PID
  nginx.service    active  true     A high performance web server     1234
```

Target all hosts to inspect the same service across the fleet:

```bash
$ osapi client node service get --target _all --name nginx.service

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  web-01
  NAME             STATUS  ENABLED  DESCRIPTION                       PID
  nginx.service    active  true     A high performance web server     1234

  web-02
  NAME             STATUS  ENABLED  DESCRIPTION                       PID
  nginx.service    active  true     A high performance web server     5678
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node service get --target web-01 \
    --name nginx.service --json
{"results":[{"hostname":"web-01","status":"ok","service":{
"name":"nginx.service","status":"active","enabled":true,
"description":"A high performance web server","pid":1234}
}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default  |
| -------------- | -------------------------------------------------------- | -------- |
| `--name`       | Service name                                             | required |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`   |
| `-j, --json`   | Output raw JSON response                                 |          |
