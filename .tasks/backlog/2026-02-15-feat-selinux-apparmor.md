---
title: SELinux/AppArmor security policy management
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add security module management. Ansible's `selinux` and `apparmor`
modules are commonly used for hardening. An appliance should expose
its security policy state.

## API Endpoints

```
GET    /security/selinux         - Get SELinux status and mode
PUT    /security/selinux         - Set SELinux mode (enforcing/permissive)
GET    /security/apparmor        - Get AppArmor status
GET    /security/apparmor/profile - List AppArmor profiles
PUT    /security/apparmor/profile/{name} - Set profile mode (enforce/complain)
```

## Operations

- `security.selinux.status.get` (query)
- `security.selinux.mode.update` (modify)
- `security.apparmor.status.get` (query)
- `security.apparmor.profiles.get` (query)
- `security.apparmor.profile.update` (modify)

## Provider

- `internal/provider/security/selinux/` — `getenforce`, `setenforce`
- `internal/provider/security/apparmor/` — `aa-status`, `aa-enforce`,
  `aa-complain`

## Notes

- Platform-specific: SELinux for RHEL, AppArmor for Ubuntu
- Provider selection based on what's installed
- Scopes: `security:read`, `security:write`
