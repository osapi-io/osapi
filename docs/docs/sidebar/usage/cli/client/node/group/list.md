# List

List all groups on a target host:

```bash
$ osapi client node group list --target web-01

  NAME     GID   MEMBERS
  docker   999   deploy,app
  sudo     27    deploy
  users    100   deploy,app
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node group list --target web-01 --json
{"results":[{"hostname":"web-01","groups":[{"name":"docker","gid":999,"members":["deploy","app"]}],"status":"ok"}],"job_id":"..."}
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `-j, --json`   | Output raw JSON response                                 |         |
