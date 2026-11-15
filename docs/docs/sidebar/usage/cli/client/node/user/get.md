# Get

Get a user account by name.

```bash
osapi client node user get
```

## Examples

```bash
osapi client node user get --target web-01 --name deploy
osapi client node user get --target web-01 --name deploy --json
```

## Flags

| Flag     | Description                         | Default |
| -------- | ----------------------------------- | ------- |
| `--name` | Name of the user account (required) |         |

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
