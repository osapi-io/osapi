# Retry

Retry a failed or stuck job:

```bash
$ osapi client job retry --job-id 550e8400-e29b-41d4-a716-446655440000

  Job ID: 660e8400-e29b-41d4-a716-446655440000    Status: created
```

## Flags

| Flag                | Description             | Default  |
| ------------------- | ----------------------- | -------- |
| `--job-id`          | Job ID to retry         | required |
| `--target-hostname` | Override target routing | `_any`   |

Creates a new job using the same operation data as the original. The original
job is preserved for history, and a "retried" event is added to its timeline
linking to the new job.
