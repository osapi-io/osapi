# Export

Export all audit log entries to a file for long-term retention. Fetches all
entries via the REST API export endpoint and writes each entry as a JSON line
(JSONL format). Requires `audit:read` permission (admin role by default).

```bash
$ osapi client audit export --output audit.jsonl

  Exported: 142    Total: 142
  Output: audit.jsonl
```

## Options

| Flag       | Default | Description                     |
| ---------- | ------- | ------------------------------- |
| `--output` | —       | Output file path (**required**) |
| `--type`   | `file`  | Export backend type             |

## Output format

The file exporter writes one JSON object per line (JSONL). Each line is a
complete audit entry:

```jsonl
{"id":"550e8400-e29b-41d4-a716-446655440000","timestamp":"2026-02-21T10:30:00Z","user":"ops@example.com","roles":["admin"],"method":"GET","path":"/node/hostname","source_ip":"127.0.0.1","response_code":200,"duration_ms":42}
{"id":"661f1234-e29b-41d4-a716-446655440111","timestamp":"2026-02-21T10:29:55Z","user":"ops@example.com","roles":["admin"],"method":"POST","path":"/job","source_ip":"127.0.0.1","response_code":201,"duration_ms":15}
```

JSONL is easy to process with standard tools:

```bash
# Count entries
wc -l audit.jsonl

# Filter by user
grep '"user":"ops@example.com"' audit.jsonl

# Pretty-print with jq
cat audit.jsonl | jq .
```
