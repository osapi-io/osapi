# Update

Update a user's password.

The client sends the password to the controller in plaintext over TLS. The
controller hashes it before it stores the job. The agent and the provider see
only the hash. osapi never writes the plaintext to disk or to logs.

```bash
osapi client node user password update
```

## Examples

```bash
osapi client node user password update --target web-01 \
  --name deploy --password 'newpass123'
osapi client node user password update --target _all \
  --name deploy --password 'newpass123'
```

## Flags

| Flag         | Description                                                       | Default |
| ------------ | ----------------------------------------------------------------- | ------- |
| `--name`     | Username to change password for (required)                        |         |
| `--password` | New password (hashed by the controller before storage) (required) |         |

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
