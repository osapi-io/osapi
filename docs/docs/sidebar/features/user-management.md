---
sidebar_position: 14
---

# User & group management

osapi manages local user accounts and groups on target hosts. It can create,
read, update and delete users and groups. It can also set passwords and lock
accounts.

## How it works

The user and group provider reads and changes the local user database of the
system: `/etc/passwd`, `/etc/shadow` and `/etc/group`. It uses the standard
system tools: `useradd`, `usermod`, `userdel`, `groupadd`, `groupmod` and
`groupdel`. All operations run on the agent and use the job system.

### Users

User operations manage local accounts:

- **List**: list all user accounts that are not system accounts
- **Get**: get one user by name
- **Create**: add a new user, with an optional UID, GID, home, shell, groups and
  password
- **Update**: change the shell, home or groups, or lock or unlock an account
- **Delete**: remove a user account
- **Password**: change the password of a user (plaintext input, hashed before
  osapi stores the job)

### SSH keys

SSH key operations manage the `~/.ssh/authorized_keys` file for a given user:

- **ListKeys**: list all authorized keys, with the type, fingerprint and comment
- **AddKey**: add a public key to the end of the authorized_keys file. It is
  idempotent: the agent does not add a key that is already there
- **RemoveKey**: remove a key by its SHA256 fingerprint

The provider reads and writes the user's `~/.ssh/authorized_keys` file directly.
It creates the `~/.ssh` directory and `authorized_keys` file with correct
permissions (`700` and `600`) if they do not exist.

### Groups

Group operations manage local groups:

- **List**: list all groups that are not system groups
- **Get**: get one group by name
- **Create**: add a new group, with an optional GID
- **Update**: set the member list of the group
- **Delete**: remove a group

## CLI usage

### Users

```bash
# List all users
$ osapi client node user list --target web-01

# Get a specific user
$ osapi client node user get --target web-01 --name deploy

# Create a user
$ osapi client node user create --target web-01 \
    --name deploy --shell /bin/bash --groups sudo,docker

# Update a user (lock account)
$ osapi client node user update --target web-01 --name deploy --lock

# Change password
$ osapi client node user set-password --target web-01 \
    --name deploy --password 'newpass123'

# Delete a user
$ osapi client node user delete --target web-01 --name deploy
```

### SSH keys

```bash
# List SSH keys for a user
$ osapi client node user ssh-key list --target web-01 --name deploy

# Add an SSH key
$ osapi client node user ssh-key create --target web-01 \
    --name deploy --key 'ssh-ed25519 AAAA... user@laptop'

# Remove an SSH key by fingerprint
$ osapi client node user ssh-key delete --target web-01 \
    --name deploy --fingerprint 'SHA256:abc123...'
```

### Groups

```bash
# List all groups
$ osapi client node group list --target web-01

# Get a specific group
$ osapi client node group get --target web-01 --name docker

# Create a group
$ osapi client node group create --target web-01 --name deploy

# Update group members
$ osapi client node group update --target web-01 \
    --name deploy --members alice,bob

# Delete a group
$ osapi client node group delete --target web-01 --name deploy
```

## Broadcast support

All user and group operations accept broadcast targets:

```bash
# List users on all hosts
$ osapi client node user list --target _all

# Create a deploy user on all web servers
$ osapi client node user create --target group:web \
    --name deploy --shell /bin/bash
```

## Permissions

| Operation          | Permission   |
| ------------------ | ------------ |
| User list/get      | `user:read`  |
| User mutations     | `user:write` |
| SSH key list       | `user:read`  |
| SSH key add/remove | `user:write` |
| Group list/get     | `user:read`  |
| Group mutations    | `user:write` |

## Supported platforms

| OS Family | Support |
| --------- | ------- |
| Debian    | Full    |
| Darwin    | Skipped |
| Linux     | Skipped |

On a platform that is not supported, user and group operations return
`status: skipped`. They do not fail.

## Related

- [CLI Reference: User](../usage/cli/client/node/user/user.md)
- [CLI Reference: SSH Key](../usage/cli/client/node/user/ssh-key/ssh-key.mdx)
- [CLI Reference: Group](../usage/cli/client/node/group/group.md)
- [SDK: User](../sdk/client/security/user.md)
- [SDK: Group](../sdk/client/security/group.md)
