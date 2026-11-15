# Set password

Set a user's password:

```bash
$ osapi client node user set-password --target web-01 \
    --name deploy --password 'newpass123'

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  NAME
  web-01    changed  true     deploy

  1 host: 1 changed
```

The client sends the password to the controller in plaintext over TLS. The
controller hashes it before it stores the job. The agent and the provider see
only the hash. osapi never writes the plaintext to disk or to logs.

Broadcast to all hosts:

```bash
$ osapi client node user set-password --target _all \
    --name deploy --password 'newpass123'

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  NAME
  web-01    changed  true     deploy
  web-02    changed  true     deploy
  mac-01    skip

  3 hosts: 2 changed, 1 skipped

  Details:
  mac-01    unsupported platform
```

## Flags

| Flag           | Description                                                       | Default |
| -------------- | ----------------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`)          | `_all`  |
| `--name`       | Username to change password for (required)                        |         |
| `--password`   | New password (hashed by the controller before storage) (required) |         |
| `-j, --json`   | Output raw JSON response                                          |         |
