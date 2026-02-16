# Add

Add a job to the NATS job queue:

```bash
$ osapi client job add \
    --json-file operation.json \
    --target-hostname _any


  Job ID: 550e8400-e29b-41d4-a716-446655440000
  Status: unprocessed
```

## Flags

| Flag                | Description                                        |
| ------------------- | -------------------------------------------------- |
| `--json-file`       | Path to the JSON file containing operation data    |
| `--target-hostname` | Target hostname (`_any`, `_all`, or specific host) |

## Example Operation File

```json
{
  "type": "system.hostname.get",
  "data": {}
}
```

## Target Hostnames

- `_any` - Route to any available worker (load balanced)
- `_all` - Broadcast to all workers
- `hostname` - Route to a specific worker by hostname
