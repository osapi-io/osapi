# Get

Get job details and status.

```bash
osapi client job get
```

## Examples

```bash
osapi client job get --id 550e8400-e29b-41d4-a716-446655440000
```

## Flags

| Flag   | Description                   | Default |
| ------ | ----------------------------- | ------- |
| `--id` | Job ID to retrieve (required) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
