---
sidebar_position: 14
---

# User & Group Management

OSAPI manages local user accounts and groups on target hosts. It provides full
CRUD operations for both users and groups, including password management and
account locking.

## How It Works

The user and group provider reads and modifies the system's local user database
(`/etc/passwd`, `/etc/shadow`, `/etc/group`) using standard system utilities
(`useradd`, `usermod`, `userdel`, `groupadd`, `groupmod`, `groupdel`). All
operations run on the agent and are dispatched via the job system.

### Users

User operations manage local accounts:

- **List** -- enumerate all non-system user accounts
- **Get** -- retrieve a specific user by name
- **Create** -- add a new user with optional UID, GID, home, shell, groups, and
  password
- **Update** -- modify shell, home, groups, or lock/unlock an account
- **Delete** -- remove a user account
- **Password** -- change a user's password (plaintext input, hashed by the
  agent)

### SSH Keys

SSH key operations manage the `~/.ssh/authorized_keys` file for a given user:

- **ListKeys** -- enumerate all authorized keys with type, fingerprint, and
  comment
- **AddKey** -- append a public key to the authorized_keys file (idempotent --
  duplicate keys are not added)
- **RemoveKey** -- remove a key by its SHA256 fingerprint

The provider reads and writes the user's `~/.ssh/authorized_keys` file directly.
It creates the `~/.ssh` directory and `authorized_keys` file with correct
permissions (`700` and `600`) if they do not exist.

### Groups

Group operations manage local groups:

- **List** -- enumerate all non-system groups
- **Get** -- retrieve a specific group by name
- **Create** -- add a new group with optional GID
- **Update** -- set the group's member list
- **Delete** -- remove a group

## CLI Usage

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
$ osapi client node user password --target web-01 \
    --name deploy --password 'newpass123'

# Delete a user
$ osapi client node user delete --target web-01 --name deploy
```

### SSH Keys

```bash
# List SSH keys for a user
$ osapi client node user ssh-key list --target web-01 --name deploy

# Add an SSH key
$ osapi client node user ssh-key add --target web-01 \
    --name deploy --key 'ssh-ed25519 AAAA... user@laptop'

# Remove an SSH key by fingerprint
$ osapi client node user ssh-key remove --target web-01 \
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

## Broadcast Support

All user and group operations support broadcast targeting:

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

## Platform Support

| Platform | Status      |
| -------- | ----------- |
| Debian   | Supported   |
| macOS    | Unsupported |
| Linux    | Unsupported |

Unsupported platforms return a `skipped` status with an `unsupported` error
message.

## Further Reading

- [CLI Reference -- User](../usage/cli/client/node/user/user.md)
- [CLI Reference -- SSH Key](../usage/cli/client/node/user/ssh-key.md)
- [CLI Reference -- Group](../usage/cli/client/node/group/group.md)
- [SDK -- User](../sdk/client/management/user.md)
- [SDK -- Group](../sdk/client/management/group.md)
