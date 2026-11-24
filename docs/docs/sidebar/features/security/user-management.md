# User & group management

osapi manages local user accounts and groups on target hosts. It can create,
read, update and delete users and groups. It can also set passwords and lock
accounts.

## How it works

The [user behavior](../../reference/behaviors/security/user/user.md) and the
[group behavior](../../reference/behaviors/security/group/group.md) read the
local user database through gohai on each platform: `/etc/passwd`, `/etc/shadow`
and `/etc/group`. They change it on the Debian family only, with `useradd`,
`usermod`, `userdel`, `chpasswd`, `groupadd`, `gpasswd` and `groupdel`. On other
platforms, a change reports `skipped`. All operations run on the agent and use
the job system.

### Users

- **List**: list each user with its IDs, its home, its shell, its groups and its
  lock state
- **Get**: get one user by name
- **Create**: add a user with `useradd`, with an optional UID, GID, home, shell,
  groups and password. A user that the host has already does not change, and
  reports `changed: false`
- **Update**: change the shell, the home or the groups with `usermod`, or lock
  or unlock the password. The values that the user has already do not change it
- **Delete**: remove a user and its home with `userdel -r`. A user that the host
  does not have reports `changed: false`
- **Passwd**: set the password of a user with `chpasswd -e`. Each run reports
  `changed: true`

The controller hashes each password before it makes the job. Thus the job, NATS
and the agent have only the hash, never the password.

### SSH keys

The [SSH key behavior](../../reference/behaviors/security/user/sshkey/sshkey.md)
manages the `~/.ssh/authorized_keys` file of a user. Its commands are under
`osapi client node user sshkey`. It reads the keys through gohai on each
platform, and it changes them on the Debian family only.

- **List**: list each authorized key, with its type, its SHA256 fingerprint, its
  comment and its options
- **Create**: add a public key to the end of the file. A key that the file has
  already, by its fingerprint, does not change it
- **Delete**: remove a key by its SHA256 fingerprint. A key that the file does
  not have reports `changed: false`

A write makes the `~/.ssh` directory and the `authorized_keys` file if they are
missing, with modes `700` and `600`, and gives them to the user with `chown`.

### Groups

The [group behavior](../../reference/behaviors/security/group/group.md) manages
local groups. It reads `/etc/group` through gohai on each platform, and it
changes groups on the Debian family only. On other platforms, a change reports
`skipped`.

- **List**: list each group with its ID and its members, by name
- **Get**: get one group by name
- **Create**: add a group with `groupadd`, with an optional GID. A group that
  the host has already does not change, and reports `changed: false`
- **Update**: set the members of the group with `gpasswd -M`. The same members,
  in any order, do not change the group
- **Delete**: remove a group with `groupdel`. A group that the host does not
  have reports `changed: false`

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

# Lock the password of a user
$ osapi client node user update --target web-01 --name deploy --lock locked

# Set the password of a user
$ osapi client node user passwd --target web-01 \
    --name deploy --password 'newpass123'

# Delete a user
$ osapi client node user delete --target web-01 --name deploy
```

### SSH keys

```bash
# List SSH keys for a user
$ osapi client node user sshkey list --target web-01 --name deploy

# Add an SSH key
$ osapi client node user sshkey create --target web-01 \
    --name deploy --key 'ssh-ed25519 AAAA... user@laptop'

# Remove an SSH key by fingerprint
$ osapi client node user sshkey delete --target web-01 \
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

The [user](../../reference/behaviors/security/user/user.md),
[group](../../reference/behaviors/security/group/group.md), and
[sshkey](../../reference/behaviors/security/user/sshkey/sshkey.md) reference
pages give the permission of each method.

## Supported platforms

| OS Family | Read | Change  |
| --------- | ---- | ------- |
| Debian    | Yes  | Yes     |
| Darwin    | Yes  | Skipped |
| Linux     | Yes  | Skipped |

On a platform that is not supported, a change returns `status: skipped`. It does
not fail.

## Related

- [CLI Reference: User](../../usage/cli/client/node/user/user.mdx)
- [CLI Reference: SSH Key](../../usage/cli/client/node/user/sshkey/sshkey.mdx)
- [CLI Reference: Group](../../usage/cli/client/node/group/group.mdx)
- [SDK: User](../../sdk/security/user.md)
- [SDK: Group](../../sdk/security/group.md)
