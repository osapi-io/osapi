# Export

Show all audit log entries, with no pages.

To save them, send the JSON to a file: osapi client audit export --json &gt;
audit.json

```bash
osapi client audit export
```

## Examples

```bash
osapi client audit export
osapi client audit export --json
```

## Permission

It needs the `audit:read` permission. `osapi token roles get --name <role>`
shows what a role grants.

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
