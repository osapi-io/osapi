# Get

Get a single audit log entry by its UUID. Requires `audit:read` permission
(admin role by default).

```bash
$ osapi client audit get --audit-id 550e8400-e29b-41d4-a716-446655440000

  ID: 550e8400-e29b-41d4-a716-446655440000
  Timestamp: 2026-02-21 10:30:00
  User: ops@example.com
  Roles: admin
  Method: GET                   Path: /node/hostname
  Status: 200                   Duration: 42ms
  Source IP: 127.0.0.1
```

## Flags

| Flag         | Description                | Default  |
| ------------ | -------------------------- | -------- |
| `--audit-id` | Audit entry ID to retrieve | required |

Use `--json` for raw JSON output:

```bash
$ osapi client audit get --audit-id 550e8400-e29b-41d4-a716-446655440000 --json
{"entry":{"id":"550e8400-e29b-41d4-a716-446655440000","timestamp":"2026-02-21T10:30:00Z","user":"ops@example.com","roles":["admin"],"method":"GET","path":"/node/hostname","source_ip":"127.0.0.1","response_code":200,"duration_ms":42}}
```
