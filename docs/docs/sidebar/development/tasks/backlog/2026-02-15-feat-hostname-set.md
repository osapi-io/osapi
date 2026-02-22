---
title: Hostname set (complement existing get)
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add the ability to set the system hostname. Currently only `system.hostname.get`
exists. Ansible's `hostname` module is heavily used for provisioning. This is a
small but important CRUD gap.

## API Endpoints

```
PUT    /system/hostname      - Set system hostname
```

## Operations

- `system.hostname.update` (modify)

## Provider

- Extend existing `internal/provider/system/host/`
- Add `SetHostname(name string) error` to `Provider` interface
- Implementation: `hostnamectl set-hostname` via cmdexec
- Update `/etc/hostname` for persistence

## Notes

- Hostname changes may require service restarts
- Validate hostname format (RFC 1123)
- Scopes: `system:write`
- Small feature — could be done alongside any other system work
