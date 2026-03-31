# Password

Change a user's password:

```bash
$ osapi client node user password --target web-01 \
    --name deploy --password 'newpass123'

  NAME     CHANGED  STATUS
  deploy   true     ok
```

The password is sent as plaintext and hashed by the agent using the system's
default hashing algorithm.

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Username to change password for (required)               |         |
| `--password`   | New password (plaintext, hashed by the agent) (required) |         |
| `-j, --json`   | Output raw JSON response                                 |         |
