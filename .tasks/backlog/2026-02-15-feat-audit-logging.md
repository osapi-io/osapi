---
title: Audit logging for API operations
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add structured audit logging for all API operations. Essential for
security compliance and troubleshooting — who did what, when.

## API Endpoints

```
GET    /audit                - Query audit log entries
GET    /audit/{id}           - Get audit entry details
```

## Implementation

This is primarily middleware, not a new provider:

- Add Echo middleware that logs every API request with:
  - Timestamp
  - Authenticated user (from JWT claims)
  - Operation (HTTP method + path)
  - Source IP
  - Request body summary (for POST/PUT)
  - Response status code
  - Duration
- Store audit entries in NATS KV or append-only file
- Structured JSON format for machine parsing

## Notes

- Read-only via API — audit logs should be tamper-resistant
- Scopes: `audit:read` (admin only)
- Consider retention policy (auto-purge after N days)
- Consider forwarding to external syslog/SIEM
- Complement to the system log viewing feature
