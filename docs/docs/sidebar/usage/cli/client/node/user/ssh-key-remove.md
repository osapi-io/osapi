# Remove

Remove an SSH authorized key by fingerprint:

```bash
$ osapi client node user ssh-key remove --target web-01 \
    --name deploy --fingerprint 'SHA256:abc123...'

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true

  1 host: 1 changed
```

The key matching the given SHA256 fingerprint is removed from the user's
`~/.ssh/authorized_keys` file. Returns `changed: false` if the fingerprint is
not found.

When targeting all hosts:

```bash
$ osapi client node user ssh-key remove --target _all \
    --name deploy --fingerprint 'SHA256:abc123...'

  Job ID: 550e8400-e29b-41d4-a716-446655440000

  HOSTNAME  STATUS   CHANGED
  web-01    changed  true
  web-02    changed  true

  2 hosts: 2 changed
```

## Flags

| Flag            | Description                                              | Default |
| --------------- | -------------------------------------------------------- | ------- |
| `-T, --target`  | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`        | Username to remove SSH key from (required)               |         |
| `--fingerprint` | SHA256 fingerprint of the key to remove (required)       |         |
| `-j, --json`    | Output raw JSON response                                 |         |
