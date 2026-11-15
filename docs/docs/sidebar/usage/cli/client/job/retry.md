# Retry

Retry a failed or stuck job:

```bash
$ osapi client job retry --id 550e8400-e29b-41d4-a716-446655440000

  Job ID: 660e8400-e29b-41d4-a716-446655440000    Status: created
```

## Flags

| Flag       | Description             | Default  |
| ---------- | ----------------------- | -------- |
| `--id`     | Job ID to retry         | required |
| `--target` | Override target routing | `_any`   |

The command makes a new job with the same operation data as the original job.
The original job stays in the history. Its timeline gets a "retried" event that
links to the new job.
