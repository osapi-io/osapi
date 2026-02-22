---
title: Audit logging for API operations
status: done
created: 2026-02-15
updated: 2026-02-21
---

## Objective

Add structured audit logging for all API operations. Essential for security
compliance and troubleshooting — who did what, when.

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

## Outcome

Implemented in `feat/audit-logging` branch. Includes:

- `audit:read` permission (admin role only)
- NATS KV-backed audit store with 30-day TTL
- Echo-level middleware (async writes, excludes health/metrics)
- OpenAPI spec with validation, generated handler + client code
- CLI commands: `osapi client audit list`, `osapi client audit get`
- Full test coverage: unit, integration (validation + RBAC)
- Documentation: CLI docs, architecture table, config reference
- CLAUDE.md updated with validation and integration test guidance

## Notes

- Read-only via API — audit logs should be tamper-resistant
- Scopes: `audit:read` (admin only)
- 30-day retention via NATS KV TTL
- No request/response bodies logged to avoid sensitive data leakage
