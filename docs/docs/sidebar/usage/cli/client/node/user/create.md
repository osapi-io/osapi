# Create

Create a new user account:

```bash
$ osapi client node user create --target web-01 \
    --name deploy --shell /bin/bash --groups sudo,docker

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  NAME
  web-01    changed  true     deploy

  1 host: 1 changed
```

Broadcast to all hosts:

```bash
$ osapi client node user create --target _all \
    --name deploy --shell /bin/bash --groups sudo,docker

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

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Username for the new account (required)                  |         |
| `--uid`        | Numeric user ID (system assigns if omitted)              |         |
| `--gid`        | Primary group ID (system assigns if omitted)             |         |
| `--home`       | Home directory path                                      |         |
| `--shell`      | Login shell path                                         |         |
| `--groups`     | Supplementary groups (comma-separated)                   |         |
| `--password`   | Initial password (plaintext, hashed by the agent)        |         |
| `--system`     | Create a system account                                  | `false` |
| `-j, --json`   | Output raw JSON response                                 |         |
