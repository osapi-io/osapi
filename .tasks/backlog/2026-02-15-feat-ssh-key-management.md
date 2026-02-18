---
title: "Feature: SSH key and access management"
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add SSH authorized key management. Appliances typically use SSH for
emergency access, and managing keys programmatically avoids manual
file editing.

## API Endpoints

```
GET    /ssh/key/{user}       - List authorized keys for user
POST   /ssh/key/{user}       - Add authorized key
DELETE /ssh/key/{user}/{fingerprint} - Remove authorized key

GET    /ssh/config           - Get SSH server config summary
PUT    /ssh/config           - Update SSH server settings
```

## Operations

- `ssh.key.list.get` (query)
- `ssh.key.add.execute`, `ssh.key.remove.execute` (modify)
- `ssh.config.get` (query)
- `ssh.config.update` (modify)

## Provider

- `internal/provider/security/ssh/`
- Manage `~/.ssh/authorized_keys` files
- Parse and manage `/etc/ssh/sshd_config` (port, auth methods, etc.)

## Notes

- SSH key changes are security-sensitive
- Scopes: `ssh:read`, `ssh:write`
- sshd_config changes require service restart — coordinate with
  service management feature
