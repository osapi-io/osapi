# Create

Create a scheduled entry.

If you do not give `--user`, the value is `root`. If the object is a Go
template, use `--content-type template`. Then the agent renders it with its
facts before it writes the file to disk. Give the values of the template with
`--var key=value`, one `--var` for each variable.

```bash
osapi client node schedule create
```

## Examples

```bash
osapi client node schedule create --target web-01 \
  --name backup-daily \
  --schedule "0 2 * * *" \
  --object backup-script \
  --user root
```

## Flags

| Flag             | Description                                              | Default |
| ---------------- | -------------------------------------------------------- | ------- |
| `--content-type` | Content type: raw or template                            |         |
| `--interval`     | Periodic interval: hourly, daily, weekly, monthly        |         |
| `--name`         | Name for the scheduled entry (required)                  |         |
| `--object`       | Name of the uploaded file in the object store (required) |         |
| `--schedule`     | Cron schedule expression, e.g. "*/5 * * * *"             |         |
| `--user`         | User to run the command as (default root)                |         |
| `--var`          | Template variable as key=value (repeatable)              |         |

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
