# Get

Get a specific group by name:

```bash
$ osapi client node group get --target web-01 --name docker

  HOSTNAME  STATUS  NAME     GID   MEMBERS
  web-01    ok      docker   999   deploy,app

  1 host: 1 ok
```

## JSON Output

```bash
$ osapi client node group get --target web-01 --name docker --json
{"results":[{"hostname":"web-01","groups":[{"name":"docker","gid":999,"members":["deploy","app"]}],"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Group name to look up (required)                         |         |
| `-j, --json`   | Output raw JSON response                                 |         |
