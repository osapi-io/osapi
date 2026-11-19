# List

List stored files.

List all files stored in the Object Store.

```bash
osapi client file list
```

## Examples

```bash
osapi client file list
osapi client file list --json
```

## Permission

It needs the `file:read` permission. `osapi token roles get --name <role>` shows
what a role grants.

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
