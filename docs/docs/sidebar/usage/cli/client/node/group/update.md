# Update

Update a group.

The `--members` flag replaces all of the member list.

```bash
osapi client node group update
```

## Examples

```bash
osapi client node group update --target web-01 \
  --name deploy --members alice,bob
```

## Flags

| Flag        | Description                                 | Default |
| ----------- | ------------------------------------------- | ------- |
| `--members` | Group members (replaces existing, required) |         |
| `--name`    | Group name to update (required)             |         |

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
