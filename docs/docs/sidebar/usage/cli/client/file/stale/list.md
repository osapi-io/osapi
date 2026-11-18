# List

List stale file deployments.

The list shows the deployments whose object changed after the last deploy. Those
are the files that you must deploy again.

```bash
osapi client file stale list
```

## Examples

```bash
osapi client file stale list
osapi client file stale list --json
```

## Permission

It needs the `file:read` permission. `osapi token roles` lists the roles that
grant it.

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../../output.md) tells how the result is printed, and what `--json`
gives.
