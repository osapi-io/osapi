# Create

Create a new user account:

```bash
$ osapi client node user create --target web-01 \
    --name deploy --shell /bin/bash --groups sudo,docker

  NAME     CHANGED  STATUS
  deploy   true     ok
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
