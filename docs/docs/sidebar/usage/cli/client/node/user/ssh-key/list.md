# List

List SSH authorized keys.

The output does not show hosts without authorized keys. It shows skipped hosts,
for example unsupported platforms, with their error.

```bash
osapi client node user ssh-key list
```

## Examples

```bash
osapi client node user ssh-key list --target web-01 --name deploy
osapi client node user ssh-key list --target _all --name deploy
osapi client node user ssh-key list --target web-01 --name deploy --json
```

## Flags

| Flag     | Description                              | Default |
| -------- | ---------------------------------------- | ------- |
| `--name` | Username to list SSH keys for (required) |         |

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
