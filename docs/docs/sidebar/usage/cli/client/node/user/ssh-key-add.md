# Add

Add an SSH authorized key for a user:

```bash
$ osapi client node user ssh-key add --target web-01 \
    --name deploy --key 'ssh-ed25519 AAAA... user@laptop'

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  CHANGED
  true
```

The key is appended to the user's `~/.ssh/authorized_keys` file. If
the file or `~/.ssh` directory does not exist, it is created with
correct permissions (`700` for the directory, `600` for the file).

Adding a key that already exists (same fingerprint) returns
`changed: false`.

When targeting all hosts:

```bash
$ osapi client node user ssh-key add --target _all \
    --name deploy --key 'ssh-ed25519 AAAA... user@laptop'

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  CHANGED
  web-01    true
  web-02    true
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Username to add SSH key for (required)                   |         |
| `--key`        | Full SSH public key line (required)                      |         |
| `-j, --json`   | Output raw JSON response                                 |         |
