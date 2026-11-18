# Retry

Retry a failed or stuck job.

The new job has the same operation as the old job. The original job stays in the
history. Its timeline gets a `retried` event that links to the new job.

```bash
osapi client job retry [flags]
```

## Examples

```bash
osapi client job retry --id 550e8400-e29b-41d4-a716-446655440000
```

## Permission

It needs the `job:write` permission. `osapi token roles` lists the roles that
grant it.

## Flags

| Flag           | Description                                                                             | Default |
| -------------- | --------------------------------------------------------------------------------------- | ------- |
| `--id`         | UUID of the job to retry (required)                                                     |         |
| `-T, --target` | The target of the retried job: _any, _all, a hostname, or a label such as group:web.dev | `_any`  |

## Global flags

| Flag               | Description                    | Default                 |
| ------------------ | ------------------------------ | ----------------------- |
| `-d, --debug`      | Enable or disable debug mode   |                         |
| `-j, --json`       | Enable JSON output             |                         |
| `-f, --osapi-file` | Path to config file            | `/etc/osapi/osapi.yaml` |
| `--url`            | URL the client will connect to | `http://0.0.0.0:8080`   |

[Output](../../output.md) tells how the result is printed, and what `--json`
gives.
