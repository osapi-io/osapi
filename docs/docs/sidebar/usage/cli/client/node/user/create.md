# Create

Create a user account.

```bash
osapi client node user create
```

## Examples

```bash
osapi client node user create --target web-01 \
  --name deploy --shell /bin/bash --groups sudo,docker
osapi client node user create --target _all \
  --name deploy --shell /bin/bash --groups sudo,docker
```

## Flags

| Flag         | Description                                                | Default |
| ------------ | ---------------------------------------------------------- | ------- |
| `--gid`      | Primary group ID (system assigns if omitted)               |         |
| `--groups`   | Supplementary groups (comma-separated)                     |         |
| `--home`     | Home directory path                                        |         |
| `--name`     | Username for the new account (required)                    |         |
| `--password` | Initial password (hashed by the controller before storage) |         |
| `--shell`    | Login shell path                                           |         |
| `--system`   | Create a system account                                    |         |
| `--uid`      | Numeric user ID (system assigns if omitted)                |         |

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
