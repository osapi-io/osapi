# Update

Update a user account.

Do not use `--lock` and `--unlock` together. Give one or more of `--shell`,
`--home`, `--groups`, `--lock` and `--unlock`.

```bash
osapi client node user update [flags]
```

## Examples

```bash
osapi client node user update --target web-01 \
  --name deploy --shell /bin/zsh
osapi client node user update --target web-01 --name deploy --lock
osapi client node user update --target web-01 --name deploy --unlock
```

## Flags

| Flag       | Description                              | Default |
| ---------- | ---------------------------------------- | ------- |
| `--groups` | Supplementary groups (replaces existing) |         |
| `--home`   | New home directory path                  |         |
| `--lock`   | Lock the account                         |         |
| `--name`   | Username to update (required)            |         |
| `--shell`  | New login shell path                     |         |
| `--unlock` | Unlock the account                       |         |

## Global flags

| Flag               | Description                                            | Default                 |
| ------------------ | ------------------------------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode                           |                         |
| `-j, --json`       | Enable JSON output                                     |                         |
| `-f, --osapi-file` | Path to config file                                    | `/etc/osapi/osapi.yaml` |
| `-T, --target`     | Target: _all, _any, hostname, or label (group:web.dev) | `_all`                  |
| `--url`            | URL the client will connect to                         | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
