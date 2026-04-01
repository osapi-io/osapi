# Delete

Delete a user account:

```bash
$ osapi client node user delete --target web-01 --name deploy

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME     CHANGED
  deploy   true
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Name of the user account to delete (required)            |         |
| `-j, --json`   | Output raw JSON response                                 |         |
