# Add

Add a job to the queue:

```bash
$ osapi client job add \
    --json-file operation.json \
    --target-hostname _any

  Job ID: 550e8400-e29b-41d4-a716-446655440000    Status: unprocessed
```

## Flags

| Flag                | Description                                        | Default  |
| ------------------- | -------------------------------------------------- | -------- |
| `--json-file`       | Path to the JSON file containing operation data    | required |
| `--target-hostname` | Target hostname (`_any`, `_all`, or specific host) | required |

## Example Operation File

```json
{
  "type": "node.hostname.get",
  "data": {}
}
```

## Target Hostnames

- `_any` - Route to any available agent (load balanced)
- `_all` - Broadcast to all agents
- `hostname` - Route to a specific agent by hostname
