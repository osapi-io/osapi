# Export

Export all audit log entries to a file, to keep them for a long time. The
command gets all entries through the export endpoint of the REST API, and writes
each entry as a JSON line (JSONL). It needs the `audit:read` permission, which
the admin role has by default.

```bash
$ osapi client audit export --output audit.jsonl

  Exported: 142    Total: 142
  Output: audit.jsonl
```

## Flags

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

Standard tools can process JSONL:

```bash
# Count entries
wc -l audit.jsonl

# Filter by user
grep '"user":"ops@example.com"' audit.jsonl

# Pretty-print with jq
cat audit.jsonl | jq .
```
