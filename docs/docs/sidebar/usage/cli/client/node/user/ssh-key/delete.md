# Delete

Delete an SSH authorized key.

The agent removes the key with the given SHA256 fingerprint from the
`~/.ssh/authorized_keys` file of the user. If it does not find the fingerprint,
the command returns `changed: false`.

```bash
osapi client node user ssh-key delete
```

## Examples

```bash
osapi client node user ssh-key delete --target web-01 \
  --name deploy --fingerprint 'SHA256:abc123...'
osapi client node user ssh-key delete --target _all \
  --name deploy --fingerprint 'SHA256:abc123...'
```

## Flags

| Flag            | Description                                     | Default |
| --------------- | ----------------------------------------------- | ------- |
| `--fingerprint` | Fingerprint of the SSH key to remove (required) |         |
| `--name`        | Username to remove SSH key from (required)      |         |

## Global flags

| Flag               | Description                                            | Default                 |
| ------------------ | ------------------------------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                           |                         |
| `-j, --json`       | Enable JSON output                                     |                         |
| `-f, --osapi-file` | Path to config file                                    | `/etc/osapi/osapi.yaml` |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev) | `_all`                  |
| `--url`            | URL the client will connect to                         | `http://0.0.0.0:8080`   |

[Output](../../../../output.md) tells how the result is printed, and what
`--json` gives.
