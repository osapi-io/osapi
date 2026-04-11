# Create

Create a new group:

```bash
$ osapi client node group create --target web-01 --name deploy

  HOSTNAME  STATUS   NAME     CHANGED
  web-01    changed  deploy   true

  1 host: 1 changed
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Group name (required)                                    |         |
| `--gid`        | Numeric group ID (system assigns if omitted)             |         |
| `--system`     | Create a system group                                    | `false` |
| `-j, --json`   | Output raw JSON response                                 |         |
