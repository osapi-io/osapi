# Get

Get job details and status from the NATS KV store:

```bash
$ osapi client job get --job-id 550e8400-e29b-41d4-a716-446655440000


  Job ID: 550e8400-e29b-41d4-a716-446655440000
  Status: completed
  Created: 2025-06-14T10:00:00Z
  Hostname: worker-node-1
  Operation: system.hostname.get
```

## Flags

| Flag       | Description        |
| ---------- | ------------------ |
| `--job-id` | Job ID to retrieve |

Job status is computed in real-time from append-only status events in the KV
store.
