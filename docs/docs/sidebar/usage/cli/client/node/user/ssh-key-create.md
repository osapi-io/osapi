# Create

Create an SSH authorized key for a user:

```bash
$ osapi client node user ssh-key create --target web-01 \
    --name deploy --key 'ssh-ed25519 AAAA... user@laptop'

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true

  1 host: 1 changed
```

The agent adds the key to the end of the `~/.ssh/authorized_keys` file of the
user. If the file or the `~/.ssh` directory does not exist, the agent makes it.
The permissions are `700` for the directory and `600` for the file.

If the key already exists (the same fingerprint), the command returns
`changed: false`.

When you target all hosts:

```bash
$ osapi client node user ssh-key create --target _all \
    --name deploy --key 'ssh-ed25519 AAAA... user@laptop'

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true
  web-02    changed  true

  2 hosts: 2 changed
```

## Flags

| Flag           | Description                                              | Default |
| -------------- | -------------------------------------------------------- | ------- |
| `-T, --target` | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`       | Username to add SSH key for (required)                   |         |
| `--key`        | Full SSH public key line (required)                      |         |
| `-j, --json`   | Output raw JSON response                                 |         |
