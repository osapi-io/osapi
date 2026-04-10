# Password

Change a user's password:

```bash
$ osapi client node user password --target web-01 \
    --name deploy --password 'newpass123'

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED  ERROR  NAME
  web-01    ok      true            deploy
```

The password is sent as plaintext and hashed by the agent using the system's
default hashing algorithm.

Broadcast to all hosts:

```bash
$ osapi client node user password --target _all \
    --name deploy --password 'newpass123'

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS  CHANGED  ERROR                 NAME
  web-01    ok      true                            deploy
  web-02    ok      true                            deploy
  mac-01    skipped false    unsupported platform
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Username to change password for (required)               |         |
| `--password`   | New password (plaintext, hashed by the agent) (required) |         |
| `-j, --json`   | Output raw JSON response                                 |         |
