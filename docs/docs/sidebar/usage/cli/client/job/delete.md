# Delete

Delete a job:

```bash
$ osapi client job delete --id 550e8400-e29b-41d4-a716-446655440000

  Job ID: 550e8400-e29b-41d4-a716-446655440000    Status: Deleted
```

## Flags

| Flag   | Description      | Default  |
| ------ | ---------------- | -------- |
| `--id` | Job ID to delete | required |

This permanently removes the job from storage.
