# Delete

Delete a container.

```bash
osapi client node container docker delete
```

## Examples

```bash
osapi client node container docker delete --id a1b2c3d4e5f6
osapi client node container docker delete --id my-nginx --force
osapi client node container docker delete --id my-nginx --target web-01
```

## Flags

| Flag      | Description                          | Default |
| --------- | ------------------------------------ | ------- |
| `--force` | Force removal of a running container |         |
| `--id`    | Container ID to remove (required)    |         |

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
