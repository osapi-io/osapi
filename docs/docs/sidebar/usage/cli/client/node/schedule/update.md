# Update

Update a scheduled entry.

The command updates only the fields that you give. If nothing changed, it
returns `Changed: false`.

```bash
osapi client node schedule update [flags]
```

## Examples

```bash
osapi client node schedule update --target web-01 \
  --name backup-daily \
  --schedule "0 3 * * *"
```

## Result

The agent does not keep the `--var` values of the last deploy. Give each
`--var key=value` again, or the template renders without them.

## Flags

| Flag             | Description                                      | Default |
| ---------------- | ------------------------------------------------ | ------- |
| `--content-type` | Content type: raw or template                    |         |
| `--name`         | Name of the scheduled entry to update (required) |         |
| `--object`       | New object to deploy                             |         |
| `--schedule`     | New cron schedule expression                     |         |
| `--user`         | New user to run the command as                   |         |
| `--var`          | Template variable as key=value (repeatable)      |         |

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
