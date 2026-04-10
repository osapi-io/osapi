# Update

Update a group's member list:

```bash
$ osapi client node group update --target web-01 \
    --name deploy --members alice,bob

  HOSTNAME  NAME     CHANGED  STATUS
  web-01    deploy   true     ok
```

The `--members` flag replaces the existing member list entirely.

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Group name to update (required)                          |         |
| `--members`    | Group members (comma-separated, replaces existing)       |         |
| `-j, --json`   | Output raw JSON response                                 |         |
