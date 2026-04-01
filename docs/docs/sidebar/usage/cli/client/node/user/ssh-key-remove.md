# Remove

Remove an SSH authorized key by fingerprint:

```bash
$ osapi client node user ssh-key remove --target web-01 \
    --name deploy --fingerprint 'SHA256:abc123...'

  HOSTNAME  CHANGED  STATUS
  web-01    true     ok
```

The key matching the given SHA256 fingerprint is removed from the user's
`~/.ssh/authorized_keys` file. Returns `changed: false` if the fingerprint is
not found.

## Flags

| Flag            | Description                                              | Default |
| --------------- | -------------------------------------------------------- | ------- |
| `-T, --target`  | Target: `_any`, `_all`, hostname, or label (`group:web`) | `_all`  |
| `--name`        | Username to remove SSH key from (required)               |         |
| `--fingerprint` | SHA256 fingerprint of the key to remove (required)       |         |
| `-j, --json`    | Output raw JSON response                                 |         |
