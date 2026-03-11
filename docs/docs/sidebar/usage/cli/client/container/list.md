# List

List containers on the target node:

```bash
$ osapi client container list

  server1
  ID            NAME        IMAGE          STATE    CREATED
  a1b2c3d4e5f6  my-nginx    nginx:latest   running  2024-01-15T10:30:00Z
  f6e5d4c3b2a1  my-redis    redis:7        running  2024-01-15T09:00:00Z
  1a2b3c4d5e6f  my-alpine   alpine:latest  stopped  2024-01-14T08:00:00Z
```

Filter by state:

```bash
$ osapi client container list --state running
$ osapi client container list --state stopped
$ osapi client container list --state all
```

Limit the number of results:

```bash
$ osapi client container list --limit 5
```

Target a specific host:

```bash
$ osapi client container list --target web-01
```

## JSON Output

Use `--json` to get the full API response:

```bash
$ osapi client container list --json
```

## Flags

| Flag           | Description                                              | Default   |
| -------------- | -------------------------------------------------------- | --------- |
| `--state`      | Filter by state: `running`, `stopped`, `all`             | `running` |
| `--limit`      | Maximum number of containers to return                   | `0`       |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_any`    |
| `-j, --json`   | Output raw JSON response                                 |           |
