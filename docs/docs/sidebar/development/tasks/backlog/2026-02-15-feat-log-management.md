---
title: Log viewing and management
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add log viewing endpoints. Appliance operators need to inspect system and
service logs for troubleshooting without SSH access.

## API Endpoints

```
GET    /log/journal          - Query systemd journal entries
GET    /log/journal/unit/{name} - Get logs for specific unit
GET    /log/syslog           - Get recent syslog entries
```

## Operations

- `log.journal.get` (query)
- `log.journal.unit.get` (query)
- `log.syslog.get` (query)

## Provider

- `internal/provider/system/log/`
- Use `journalctl` with JSON output for structured log parsing
- Support query params: since, until, unit, priority, limit, grep
- Return type: `LogEntry` with timestamp, unit, priority, message, PID, hostname

## Notes

- Logs can be very large — pagination and limits are essential
- Support streaming in future (SSE or WebSocket) for tail -f equivalent
- Read-only — no log deletion via API
- Scopes: `log:read`
- Consider security: some logs may contain sensitive info
