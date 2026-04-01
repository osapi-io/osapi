# Update

Update an existing user account:

```bash
$ osapi client node user update --target web-01 \
    --name deploy --shell /bin/zsh

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  NAME     CHANGED
  deploy   true
```

Lock or unlock an account:

```bash
$ osapi client node user update --target web-01 --name deploy --lock
$ osapi client node user update --target web-01 --name deploy --unlock
```

The `--lock` and `--unlock` flags are mutually exclusive. At least one of
`--shell`, `--home`, `--groups`, `--lock`, or `--unlock` must be specified.

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
