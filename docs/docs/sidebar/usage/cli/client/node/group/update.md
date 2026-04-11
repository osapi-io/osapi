# Update

Update a group's member list:

```bash
$ osapi client node group update --target web-01 \
    --name deploy --members alice,bob

  HOSTNAME  STATUS   NAME     CHANGED
  web-01    changed  deploy   true

  1 host: 1 changed
```

The `--members` flag replaces the existing member list entirely.

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Group name to update (required)                          |         |
| `--members`    | Group members (comma-separated, replaces existing)       |         |
| `-j, --json`   | Output raw JSON response                                 |         |
