# Delete

Delete a job from the NATS KV store:

```bash
$ osapi client job delete --job-id 550e8400-e29b-41d4-a716-446655440000


  Job ID: 550e8400-e29b-41d4-a716-446655440000
  Status: Deleted
```

## Flags

| Flag       | Description      | Default  |
| ---------- | ---------------- | -------- |
| `--job-id` | Job ID to delete | required |

This permanently removes the job from storage.
