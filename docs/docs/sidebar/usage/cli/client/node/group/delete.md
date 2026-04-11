# Delete

Delete a group:

```bash
$ osapi client node group delete --target web-01 --name deploy

  HOSTNAME  STATUS   NAME     CHANGED
  web-01    changed  deploy   true

  1 host: 1 changed
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Name of the group to delete (required)                   |         |
| `-j, --json`   | Output raw JSON response                                 |         |
