# Password

Change a user's password:

```bash
$ osapi client node user password --target web-01 \
    --name deploy --password 'newpass123'

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  NAME
  web-01    changed  true     deploy

  1 host: 1 changed
```

The password is sent to the controller in plaintext over TLS. The controller
hashes it before the job is stored; the agent and provider only ever see the
hash, and the plaintext is never written to disk or logs.

Broadcast to all hosts:

```bash
$ osapi client node user password --target _all \
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
