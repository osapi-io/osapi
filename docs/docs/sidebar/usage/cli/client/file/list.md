# List

List stored files.

With `--stale`, the list shows the deployments whose object changed after the
last deploy. Those are the files that you must deploy again.

```bash
osapi client file list [flags]
```

## Examples

```bash
osapi client file list
osapi client file list --json
osapi client file list --stale
```

## Flags

| Flag      | Description                                                       | Default |
| --------- | ----------------------------------------------------------------- | ------- |
| `--stale` | List only deployments whose file changed since they were deployed |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
