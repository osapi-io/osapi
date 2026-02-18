---
title: System update and patch management
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add OS-level update and patching endpoints. Keeping an appliance patched
is critical for security. This is related to but distinct from package
management — it focuses on the overall system update lifecycle.

## API Endpoints

```
GET    /update/status        - Check for available updates (security, all)
POST   /update/check         - Trigger update check
POST   /update/apply         - Apply available updates
GET    /update/history       - List past update operations
POST   /update/rollback      - Rollback last update (if supported)
```

## Operations

- `update.status.get`, `update.history.get` (query)
- `update.check.execute`, `update.apply.execute` (modify)
- `update.rollback.execute` (modify)

## Provider

- `internal/provider/system/update/`
- Implementations: `apt_provider.go` (`apt update`, `apt upgrade`),
  `yum_provider.go` (`yum check-update`, `yum update`)
- Return types: `UpdateStatus` with available count, security count,
  last checked timestamp; `UpdateRecord` with date, packages, success

## Notes

- Updates are long-running — perfect for async job system
- Security updates vs full updates should be distinguishable
- Scopes: `update:read`, `update:write`
- Consider auto-update configuration (enable/disable, schedule)
- Reboot may be required after kernel updates — coordinate with
  power management feature
