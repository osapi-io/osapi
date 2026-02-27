---
title: Package management (apt/yum)
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add package management endpoints for querying and managing installed packages.
Essential for maintaining system software and security patches.

## API Endpoints

```
GET    /package               - List installed packages (with filters)
GET    /package/{name}        - Get package details
POST   /package/{name}/install  - Install a package
POST   /package/{name}/remove   - Remove a package
POST   /package/update         - Update all packages (apt upgrade)
GET    /package/updates        - List available updates
```

## Operations

- `package.list.get` (query)
- `package.status.get` (query)
- `package.updates.get` (query)
- `package.install.execute` (modify)
- `package.remove.execute` (modify)
- `package.update.execute` (modify)

## Provider

- `internal/provider/node/package/`
- Implementations: `apt_provider.go` (Debian/Ubuntu), `yum_provider.go`
  (RHEL/CentOS)
- Return type: `PackageInfo` with name, version, description, installed size,
  status

## Notes

- Package operations are long-running — good fit for async job system
- Consider security: only allow packages from configured repositories
- Scopes: `package:read`, `package:write`
- `package/update` (upgrade all) should require elevated scope
