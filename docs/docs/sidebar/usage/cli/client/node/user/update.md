# Update

Update a user account:

```bash
$ osapi client node user update --target web-01 \
    --name deploy --shell /bin/zsh

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED  NAME
  web-01    changed  true     deploy

  1 host: 1 changed
```

Lock or unlock an account:

```bash
$ osapi client node user update --target web-01 --name deploy --lock
$ osapi client node user update --target web-01 --name deploy --unlock
```

Do not use `--lock` and `--unlock` together. Give one or more of `--shell`,
`--home`, `--groups`, `--lock` and `--unlock`.

Broadcast to all hosts:

```bash
$ osapi client node user update --target _all \
    --name deploy --shell /bin/zsh

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
| `--name`       | Username to update (required)                            |         |
| `--shell`      | New login shell path                                     |         |
| `--home`       | New home directory path                                  |         |
| `--groups`     | Supplementary groups (replaces existing)                 |         |
| `--lock`       | Lock the account                                         | `false` |
| `--unlock`     | Unlock the account                                       | `false` |
| `-j, --json`   | Output raw JSON response                                 |         |
