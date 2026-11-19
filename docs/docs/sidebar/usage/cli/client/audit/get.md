# Get

Get a single audit log entry.

Returns a single audit log entry by ID.

```bash
osapi client audit get [flags]
```

## Examples

```bash
osapi client audit get --id 550e8400-e29b-41d4-a716-446655440000
osapi client audit get --id 550e8400-e29b-41d4-a716-446655440000 --json
```

## Permission

It needs the `audit:read` permission. `osapi token roles list` lists the roles
that grant it.

## Flags

| Flag   | Description                   | Default |
| ------ | ----------------------------- | ------- |
| `--id` | The audit entry ID (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
