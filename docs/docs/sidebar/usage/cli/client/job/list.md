# List

List jobs.

```bash
osapi client job list [flags]
```

## Examples

```bash
osapi client job list
osapi client job list --status completed --limit 3
```

## Flags

| Flag       | Description                                                                       | Default |
| ---------- | --------------------------------------------------------------------------------- | ------- |
| `--limit`  | Maximum number of jobs per page (1-100, default 10)                               | `10`    |
| `--offset` | Skip the first N jobs (for pagination)                                            |         |
| `--status` | Filter jobs by status (submitted, processing, completed, failed, partial_failure) |         |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
