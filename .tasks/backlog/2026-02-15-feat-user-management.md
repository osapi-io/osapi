---
title: User and group management
status: backlog
created: 2026-02-15
updated: 2026-02-15
---

## Objective

Add user and group management endpoints. An appliance needs to manage
local system accounts for access control and auditing.

## API Endpoints

```
GET    /user                - List system users (with filters)
GET    /user/{name}         - Get user details
POST   /user                - Create user
PUT    /user/{name}         - Update user (shell, groups, etc.)
DELETE /user/{name}         - Delete user
POST   /user/{name}/password - Change user password

GET    /group               - List groups
GET    /group/{name}        - Get group details
POST   /group               - Create group
PUT    /group/{name}        - Update group membership
DELETE /group/{name}        - Delete group
```

## Operations

- `user.list.get`, `user.status.get` (query)
- `user.create.execute`, `user.update.execute`,
  `user.delete.execute`, `user.password.execute` (modify)
- `group.list.get`, `group.status.get` (query)
- `group.create.execute`, `group.update.execute`,
  `group.delete.execute` (modify)

## Provider

- `internal/provider/system/user/`
- Parse `/etc/passwd`, `/etc/shadow`, `/etc/group`
- Implementations: `useradd`, `usermod`, `userdel`, `groupadd`, etc.
- Return type: `UserInfo` with name, UID, GID, home, shell, groups,
  locked status, last login

## Notes

- Never expose password hashes via API
- Password changes should require current password or admin scope
- Scopes: `user:read`, `user:write`
- Consider filtering out system users (UID < 1000) by default
