# List

List containers on the target node:

```bash
$ osapi client node container docker list

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  NAME        IMAGE          STATE
  server1   ok      my-nginx    nginx:latest   running
  server1   ok      my-redis    redis:7        running
  server1   ok      my-alpine   alpine:latest  stopped

  1 host: 1 ok
```

Filter by state:

```bash
$ osapi client node container docker list --state running
$ osapi client node container docker list --state stopped
$ osapi client node container docker list --state all
```

Limit the number of results:

```bash
$ osapi client node container docker list --limit 5
```

Target a specific host:

```bash
$ osapi client node container docker list --target web-01
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client node container docker list --json
```

## Flags

| Flag           | Description                                              | Default   |
| -------------- | -------------------------------------------------------- | --------- |
| `--state`      | Filter by state: `running`, `stopped`, `all`             | `running` |
| `--limit`      | Maximum number of containers to return                   | `0`       |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`    |
| `-j, --json`   | Output raw JSON response                                 |           |
