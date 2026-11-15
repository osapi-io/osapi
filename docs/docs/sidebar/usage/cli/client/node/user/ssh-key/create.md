# Create

Create an SSH authorized key.

The agent adds the key to the end of the `~/.ssh/authorized_keys` file of the
user. If the file or the `~/.ssh` directory does not exist, the agent makes it.
The permissions are `700` for the directory and `600` for the file.

```bash
osapi client node user ssh-key create
```

## Examples

```bash
osapi client node user ssh-key create --target web-01 \
  --name deploy --key 'ssh-ed25519 AAAA... user@laptop'
osapi client node user ssh-key create --target _all \
  --name deploy --key 'ssh-ed25519 AAAA... user@laptop'
```

## Result

If the key already exists (the same fingerprint), the command returns
`changed: false`.

## Flags

| Flag     | Description                            | Default |
| -------- | -------------------------------------- | ------- |
| `--key`  | Full SSH public key line (required)    |         |
| `--name` | Username to add SSH key for (required) |         |

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
