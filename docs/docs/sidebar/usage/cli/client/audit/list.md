# List

List audit log entries with pagination. Shows recent API activity including
user, HTTP method, path, response status, and duration. Requires `audit:read`
permission (admin role by default).

```bash
$ osapi client audit list

  Total: 3

  Audit Entries:

  ┏━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━━━━┓
  ┃ ID       ┃ TIMESTAMP           ┃ USER            ┃ METHOD ┃ PATH              ┃ STATUS ┃ DURATION ┃
  ┣━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━╋━━━━━━━━╋━━━━━━━━━━━━━━━━━━━╋━━━━━━━━╋━━━━━━━━━━┫
  ┃ 550e…000 ┃ 2026-02-21 10:30:00 ┃ ops@example.com ┃ GET    ┃ /node/hostname  ┃ 200    ┃ 42ms     ┃
  ┃ 661f…111 ┃ 2026-02-21 10:29:55 ┃ ops@example.com ┃ POST   ┃ /job              ┃ 201    ┃ 15ms     ┃
  ┃ 772a…222 ┃ 2026-02-21 10:29:50 ┃ ops@example.com ┃ GET    ┃ /network/dns/eth0 ┃ 200    ┃ 8ms      ┃
  ┗━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━┻━━━━━━━━┻━━━━━━━━━━━━━━━━━━━┻━━━━━━━━┻━━━━━━━━━━┛
```

Use `--limit` and `--offset` for pagination:

```bash
$ osapi client audit list --limit 10 --offset 20
```

Use `--json` for raw JSON output:

```bash
$ osapi client audit list --json
{"total_items":3,"items":[{"id":"550e8400-e29b-41d4-a716-446655440000","timestamp":"2026-02-21T10:30:00Z","user":"ops@example.com","roles":["admin"],"method":"GET","path":"/node/hostname","source_ip":"127.0.0.1","response_code":200,"duration_ms":42}]}
```
