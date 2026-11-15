# Export

Export audit log entries to a file.

The command gets all entries through the export endpoint of the REST API, and
writes each entry as a JSON line (JSONL).

```bash
osapi client audit export [flags]
```

## Examples

```bash
osapi client audit export --output audit.jsonl
```

## Permission

It needs the `audit:read` permission. `osapi token roles` lists the roles that
grant it.

## Flags

| Flag       | Description                 | Default |
| ---------- | --------------------------- | ------- |
| `--output` | Output file path (required) |         |
| `--type`   | Export backend type         | `file`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
