# Export

Export all audit log entries to a file for long-term retention. Paginates
through all entries via the REST API and writes each entry as a JSON line
(JSONL format). Requires `audit:read` permission (admin role by default).

```bash
$ osapi client audit export --output audit.jsonl

  Exported: 142    Total: 142
  Output: audit.jsonl
```

## Options

| Flag           | Default | Description                     |
| -------------- | ------- | ------------------------------- |
| `--output`     | —       | Output file path (**required**) |
| `--type`       | `file`  | Export backend type             |
| `--batch-size` | `100`   | Number of entries per API call  |

## Custom batch size

Use `--batch-size` to control how many entries are fetched per API call.
Larger values reduce the number of requests but use more memory per batch:

```bash
$ osapi client audit export --output audit.jsonl --batch-size 50
```

## Output format

The file exporter writes one JSON object per line (JSONL). Each line is a
complete audit entry:

```jsonl
{"id":"550e8400-e29b-41d4-a716-446655440000","timestamp":"2026-02-21T10:30:00Z","user":"ops@example.com","roles":["admin"],"method":"GET","path":"/system/hostname","source_ip":"127.0.0.1","response_code":200,"duration_ms":42}
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
