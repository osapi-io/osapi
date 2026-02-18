# Get

Get job details and status:

```bash
$ osapi client job get --job-id 550e8400-e29b-41d4-a716-446655440000


  Job ID: 550e8400-e29b-41d4-a716-446655440000
  Status: completed
  Created: 2025-06-14T10:00:00Z
  Hostname: worker-node-1
  Operation: system.hostname.get
```

## Flags

| Flag       | Description        | Default  |
| ---------- | ------------------ | -------- |
| `--job-id` | Job ID to retrieve | required |

Job status is retrieved from the API and reflects the current state of the job.
