# List

List audit log entries.

The list shows recent API activity: the user, the HTTP method, the path, the
response status and the duration.

```bash
osapi client audit list [flags]
```

## Examples

```bash
osapi client audit list
osapi client audit list --limit 10 --offset 20
osapi client audit list --json
```

## Permission

It needs the `audit:read` permission. `osapi token roles list` lists the roles
that grant it.

## Flags

| Flag       | Description                         | Default |
| ---------- | ----------------------------------- | ------- |
| `--limit`  | Maximum number of entries to return | `20`    |
| `--offset` | Number of entries to skip           |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
