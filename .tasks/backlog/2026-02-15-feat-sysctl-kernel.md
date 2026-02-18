---
title: "Feature: Kernel parameter management (sysctl)"
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add kernel parameter query and tuning. Appliances often need specific
sysctl settings for performance or security (e.g., IP forwarding,
connection limits, kernel hardening).

## API Endpoints

```
GET    /sysctl               - List all sysctl parameters (with filter)
GET    /sysctl/{key}         - Get specific parameter value
PUT    /sysctl/{key}         - Set parameter value (runtime)
PUT    /sysctl/{key}/persist - Set parameter value (persistent across reboot)
```

## Operations

- `sysctl.list.get`, `sysctl.status.get` (query)
- `sysctl.update` (modify — runtime only)
- `sysctl.persist.update` (modify — writes to `/etc/sysctl.d/`)

## Provider

- `internal/provider/system/sysctl/`
- Parse `sysctl -a` output and `/etc/sysctl.d/*.conf`
- Use `sysctl -w` for runtime changes
- Write conf files for persistent changes

## Notes

- Incorrect sysctl values can break networking or crash system
- Consider a whitelist of safe-to-modify parameters
- Scopes: `sysctl:read`, `sysctl:write`
