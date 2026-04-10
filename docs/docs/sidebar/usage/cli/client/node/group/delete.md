# Delete

Delete a group:

```bash
$ osapi client node group delete --target web-01 --name deploy

  HOSTNAME  NAME     CHANGED  STATUS
  web-01    deploy   true     ok
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Name of the group to delete (required)                   |         |
| `-j, --json`   | Output raw JSON response                                 |         |
