---
sidebar_position: 5
---

# User

User account management on target hosts.

## Methods

| Method                                                            | Returns                                  | Permission     | Description                            |
| ----------------------------------------------------------------- | ---------------------------------------- | -------------- | -------------------------------------- |
| `c.User.List(ctx, hostname)`                                      | `Response[Collection[UserAccount]]`      | `user:read`    | List the users of the host             |
| `c.User.Create(ctx, hostname, input UserCreateInput)`             | `Response[Collection[UserResult]]`       | `user:write`   | Add a user                             |
| `c.User.Get(ctx, hostname, name)`                                 | `Response[Collection[UserAccount]]`      | `user:read`    | Get one user                           |
| `c.User.Update(ctx, hostname, input UserUpdateInput)`             | `Response[Collection[UserResult]]`       | `user:write`   | Change a user                          |
| `c.User.Delete(ctx, hostname, name)`                              | `Response[Collection[UserResult]]`       | `user:write`   | Remove a user                          |
| `c.User.Passwd(ctx, hostname, input UserPasswdInput)`             | `Response[Collection[UserResult]]`       | `user:execute` | Set the password of a user             |
| `c.UserSSHKey.List(ctx, hostname, name)`                          | `Response[Collection[UserSshkeyKey]]`    | `user:read`    | List the authorized SSH keys of a user |
| `c.UserSSHKey.Create(ctx, hostname, input UserSSHKeyCreateInput)` | `Response[Collection[UserSshkeyResult]]` | `user:write`   | Add an authorized SSH key to a user    |
| `c.UserSSHKey.Delete(ctx, hostname, input UserSSHKeyDeleteInput)` | `Response[Collection[UserSshkeyResult]]` | `user:write`   | Remove an authorized SSH key of a user |

## Types

The fields of each type are the properties of its schema in the API's spec.

### UserAccount

| Field      | Type       | Description                                                                            |
| ---------- | ---------- | -------------------------------------------------------------------------------------- |
| `Error`    | `string`   | Error message if the agent failed.                                                     |
| `GID`      | `int`      | Its primary group ID.                                                                  |
| `Groups`   | `[]string` | The groups it is a member of, besides its primary group.                               |
| `Home`     | `string`   | Its home directory.                                                                    |
| `Hostname` | `string`   | Hostname of the agent that produced this result.                                       |
| `Locked`   | `bool`     | Whether a login with its password is refused.                                          |
| `Name`     | `string`   | The user.                                                                              |
| `Shell`    | `string`   | Its login shell.                                                                       |
| `Status`   | `string`   | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `UID`      | `int`      | Its numeric user ID.                                                                   |

### UserCreateInput

| Field      | Type       | Description                                                          |
| ---------- | ---------- | -------------------------------------------------------------------- |
| `GID`      | `int`      | Its primary group ID; a group of its own when it is 0.               |
| `Groups`   | `[]string` | The groups it is a member of, besides its primary group.             |
| `Home`     | `string`   | Its home directory; /home/<name> when empty.                         |
| `Name`     | `string`   | The user to add, such as alice.                                      |
| `Password` | `string`   | Its password, which osapi hashes on the controller; none when empty. |
| `Shell`    | `string`   | Its login shell, such as /bin/bash; the system's default when empty. |
| `System`   | `bool`     | Add it as a system user, with an ID from the system range.           |
| `UID`      | `int`      | Its numeric user ID; the system assigns one when it is 0.            |

### UserResult

| Field      | Type     | Description                                                                            |
| ---------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`  | `bool`   | Whether the host changed.                                                              |
| `Error`    | `string` | Error message if the agent failed.                                                     |
| `Hostname` | `string` | Hostname of the agent that produced this result.                                       |
| `Name`     | `string` | The user.                                                                              |
| `Status`   | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### UserUpdateInput

| Field    | Type       | Description                                                                                                            |
| -------- | ---------- | ---------------------------------------------------------------------------------------------------------------------- |
| `Name`   | `string`   | The name of the path: the user, such as alice.                                                                         |
| `Groups` | `[]string` | The groups it is a member of, which replace the groups it has.                                                         |
| `Home`   | `string`   | Its home directory, which usermod moves the files to.                                                                  |
| `Lock`   | `string`   | Whether to refuse logins with its password: locked refuses them, unlocked allows them. It is one of: locked, unlocked. |
| `Shell`  | `string`   | Its login shell, such as /bin/bash.                                                                                    |

### UserPasswdInput

| Field      | Type     | Description                                             |
| ---------- | -------- | ------------------------------------------------------- |
| `Name`     | `string` | The name of the path: the user, such as alice.          |
| `Password` | `string` | The new password, which osapi hashes on the controller. |

### UserSshkeyKey

| Field         | Type     | Description                                                                            |
| ------------- | -------- | -------------------------------------------------------------------------------------- |
| `Comment`     | `string` | The comment after the key, such as alice@laptop.                                       |
| `Error`       | `string` | Error message if the agent failed.                                                     |
| `Fingerprint` | `string` | The SHA256 fingerprint of the key.                                                     |
| `Hostname`    | `string` | Hostname of the agent that produced this result.                                       |
| `Options`     | `string` | The options before the key, such as from="10.0.0.0/8".                                 |
| `Status`      | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |
| `Type`        | `string` | The key type, such as ssh-ed25519.                                                     |

### UserSSHKeyCreateInput

| Field  | Type     | Description                                                                                   |
| ------ | -------- | --------------------------------------------------------------------------------------------- |
| `Name` | `string` | The name of the path: the user, such as alice.                                                |
| `Key`  | `string` | The public key line, such as ssh-ed25519 AAAA... alice@laptop, with options before it if any. |

### UserSshkeyResult

| Field         | Type     | Description                                                                            |
| ------------- | -------- | -------------------------------------------------------------------------------------- |
| `Changed`     | `bool`   | Whether the host changed.                                                              |
| `Error`       | `string` | Error message if the agent failed.                                                     |
| `Fingerprint` | `string` | The SHA256 fingerprint of the key.                                                     |
| `Hostname`    | `string` | Hostname of the agent that produced this result.                                       |
| `Status`      | `string` | The status of the operation for this host. It is one of: ok, failed, skipped, timeout. |

### UserSSHKeyDeleteInput

| Field         | Type     | Description                                          |
| ------------- | -------- | ---------------------------------------------------- |
| `Name`        | `string` | The name of the path: the user, such as alice.       |
| `Fingerprint` | `string` | The SHA256 fingerprint of the key, as list shows it. |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all users
resp, err := c.User.List(ctx, "_any")
// Each result is one user on one host.
for _, r := range resp.Data.Results {
    fmt.Printf("%s: %s uid=%d\n", r.Hostname, r.Name, r.UID)
}

// Get a specific user
resp, err := c.User.Get(ctx, "web-01", "deploy")

// Create a user
resp, err := c.User.Create(ctx, "web-01", client.UserCreateInput{
    Name:   "deploy",
    Shell:  "/bin/bash",
    Groups: []string{"sudo", "docker"},
})

// Update a user (lock account)
resp, err := c.User.Update(ctx, "web-01", client.UserUpdateInput{
    Name: "deploy",
    Lock: "locked",
})

// Set the password. The controller hashes it before it makes the job.
resp, err := c.User.Passwd(ctx, "web-01", client.UserPasswdInput{
    Name: "deploy",
    Password: "newpass123",
})

// Delete a user
resp, err := c.User.Delete(ctx, "web-01", "deploy")

// List SSH keys for a user
keysResp, err := c.UserSSHKey.List(ctx, "web-01", "deploy")
// Each result is one key on one host.
for _, r := range keysResp.Data.Results {
    fmt.Printf("%s %s %s\n", r.Type, r.Fingerprint, r.Comment)
}

// Add an SSH key
addResp, err := c.UserSSHKey.Create(ctx, "web-01", client.UserSSHKeyCreateInput{
    Name: "deploy",
    Key: "ssh-ed25519 AAAA... user@laptop",
})

// Remove an SSH key by fingerprint
removeResp, err := c.UserSSHKey.Delete(ctx, "web-01", client.UserSSHKeyDeleteInput{
    Name: "deploy",
    Fingerprint: "SHA256:abc123...",
})
```

## Example

Refer to
[`examples/sdk/client/user.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/user.go)
for a complete working example.

## Permissions

The Permission column of [Methods](#methods) gives the permission of each
method.
