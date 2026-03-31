# User & Group Management Provider Design

## Overview

Add user and group management to OSAPI. CRUD operations for local
system users and groups. Uses `useradd`, `usermod`, `userdel`,
`groupadd`, `groupmod`, `groupdel`, and `chpasswd` via
`exec.Manager`. Never exposes password hashes.

## Architecture

Direct provider at `internal/provider/node/user/`. One provider
package with two sets of methods (users and groups). Two API path
prefixes under `/node/{hostname}/`: `user/` and `group/`.

- **Category**: `node`
- **Permissions**: `user:read`, `user:write`
- **Provider type**: direct (exec.Manager)

## Provider Interface

```go
type Provider interface {
    // Users
    ListUsers(ctx context.Context) ([]User, error)
    GetUser(ctx context.Context, name string) (*User, error)
    CreateUser(ctx context.Context, opts CreateUserOpts) (*UserResult, error)
    UpdateUser(ctx context.Context, name string, opts UpdateUserOpts) (*UserResult, error)
    DeleteUser(ctx context.Context, name string) (*UserResult, error)
    ChangePassword(ctx context.Context, name string, password string) (*UserResult, error)

    // Groups
    ListGroups(ctx context.Context) ([]Group, error)
    GetGroup(ctx context.Context, name string) (*Group, error)
    CreateGroup(ctx context.Context, opts CreateGroupOpts) (*GroupResult, error)
    UpdateGroup(ctx context.Context, name string, opts UpdateGroupOpts) (*GroupResult, error)
    DeleteGroup(ctx context.Context, name string) (*GroupResult, error)
}
```

## Data Types

```go
type User struct {
    Name   string   `json:"name"`
    UID    int      `json:"uid"`
    GID    int      `json:"gid"`
    Home   string   `json:"home"`
    Shell  string   `json:"shell"`
    Groups []string `json:"groups,omitempty"`
    Locked bool     `json:"locked"`
}

type CreateUserOpts struct {
    Name     string   `json:"name"`
    UID      int      `json:"uid,omitempty"`
    GID      int      `json:"gid,omitempty"`
    Home     string   `json:"home,omitempty"`
    Shell    string   `json:"shell,omitempty"`
    Groups   []string `json:"groups,omitempty"`
    Password string   `json:"password,omitempty"`
    System   bool     `json:"system,omitempty"`
}

type UpdateUserOpts struct {
    Shell  string   `json:"shell,omitempty"`
    Home   string   `json:"home,omitempty"`
    Groups []string `json:"groups,omitempty"`
    Lock   *bool    `json:"lock,omitempty"`
}

type Group struct {
    Name    string   `json:"name"`
    GID     int      `json:"gid"`
    Members []string `json:"members,omitempty"`
}

type CreateGroupOpts struct {
    Name   string `json:"name"`
    GID    int    `json:"gid,omitempty"`
    System bool   `json:"system,omitempty"`
}

type UpdateGroupOpts struct {
    Members []string `json:"members,omitempty"`
}

type UserResult struct {
    Name    string `json:"name"`
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}

type GroupResult struct {
    Name    string `json:"name"`
    Changed bool   `json:"changed"`
    Error   string `json:"error,omitempty"`
}
```

## Debian Implementation

- **ListUsers**: parse `/etc/passwd`, filter UID >= 1000 by default
  (skip system users like root, daemon, etc.)
- **GetUser**: parse `/etc/passwd` for specific user, read groups
  via `id -Gn <name>`
- **CreateUser**: `useradd` with flags for UID (`-u`), home (`-d`),
  shell (`-s`), groups (`-G`), system (`-r`), create home (`-m`).
  If password provided, pipe to `chpasswd` after creation.
- **UpdateUser**: `usermod` with flags for shell (`-s`), home
  (`-d -m`), groups (`-G`). Lock via `usermod -L`, unlock via
  `usermod -U`.
- **DeleteUser**: `userdel -r` (removes home directory)
- **ChangePassword**: echo `name:password` and pipe to `chpasswd`
- **ListGroups**: parse `/etc/group`
- **GetGroup**: parse `/etc/group` for specific group
- **CreateGroup**: `groupadd` with optional GID (`-g`), system
  (`-r`)
- **UpdateGroup**: `groupmod` for membership. Use `gpasswd -M` to
  set member list.
- **DeleteGroup**: `groupdel`

Never expose password hashes — only read from `/etc/passwd` (which
doesn't contain hashes), not `/etc/shadow`.

## Platform Implementations

| Platform | Implementation                                  |
| -------- | ----------------------------------------------- |
| Debian   | useradd/usermod/userdel/groupadd/groupmod/groupdel |
| Darwin   | ErrUnsupported                                  |
| Linux    | ErrUnsupported                                  |

## Container Behavior

Return `ErrUnsupported` in containers — user/group management is
the host's concern.

## API Endpoints

### User Endpoints

| Method   | Path                                    | Permission   | Description      |
| -------- | --------------------------------------- | ------------ | ---------------- |
| `GET`    | `/node/{hostname}/user`                 | `user:read`  | List users       |
| `GET`    | `/node/{hostname}/user/{name}`          | `user:read`  | Get user         |
| `POST`   | `/node/{hostname}/user`                 | `user:write` | Create user      |
| `PUT`    | `/node/{hostname}/user/{name}`          | `user:write` | Update user      |
| `DELETE` | `/node/{hostname}/user/{name}`          | `user:write` | Delete user      |
| `POST`   | `/node/{hostname}/user/{name}/password` | `user:write` | Change password  |

### Group Endpoints

| Method   | Path                              | Permission   | Description    |
| -------- | --------------------------------- | ------------ | -------------- |
| `GET`    | `/node/{hostname}/group`          | `user:read`  | List groups    |
| `GET`    | `/node/{hostname}/group/{name}`   | `user:read`  | Get group      |
| `POST`   | `/node/{hostname}/group`          | `user:write` | Create group   |
| `PUT`    | `/node/{hostname}/group/{name}`   | `user:write` | Update group   |
| `DELETE` | `/node/{hostname}/group/{name}`   | `user:write` | Delete group   |

All endpoints support broadcast targeting.

## SDK

Two SDK services sharing the same permissions:

```go
// UserService
client.User.List(ctx, host)
client.User.Get(ctx, host, name)
client.User.Create(ctx, host, opts)
client.User.Update(ctx, host, name, opts)
client.User.Delete(ctx, host, name)
client.User.ChangePassword(ctx, host, name, password)

// GroupService
client.Group.List(ctx, host)
client.Group.Get(ctx, host, name)
client.Group.Create(ctx, host, opts)
client.Group.Update(ctx, host, name, opts)
client.Group.Delete(ctx, host, name)
```

## Permissions

- `user:read` — list and get (users and groups). Added to admin,
  write, and read roles.
- `user:write` — create, update, delete, change password. Added
  to admin and write roles.
