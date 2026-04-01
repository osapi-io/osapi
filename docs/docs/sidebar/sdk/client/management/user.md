---
sidebar_position: 5
---

# User

User account management on target hosts.

## Methods

| Method                                          | Description                      |
| ----------------------------------------------- | -------------------------------- |
| `List(ctx, hostname)`                           | List all user accounts           |
| `Get(ctx, hostname, name)`                      | Get a user by name               |
| `Create(ctx, hostname, opts)`                   | Create a user account            |
| `Update(ctx, hostname, name, opts)`             | Update a user account            |
| `Delete(ctx, hostname, name)`                   | Delete a user account            |
| `ChangePassword(ctx, hostname, name, password)` | Change a user's password         |
| `ListKeys(ctx, hostname, name)`                 | List SSH authorized keys         |
| `AddKey(ctx, hostname, name, opts)`             | Add an SSH authorized key        |
| `RemoveKey(ctx, hostname, name, fingerprint)`   | Remove an SSH key by fingerprint |

## Usage

```go
import "github.com/retr0h/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all users
resp, err := c.User.List(ctx, "_any")
for _, r := range resp.Data.Results {
    for _, u := range r.Users {
        fmt.Printf("%s uid=%d\n", u.Name, u.UID)
    }
}

// Get a specific user
resp, err := c.User.Get(ctx, "web-01", "deploy")

// Create a user
resp, err := c.User.Create(ctx, "web-01", client.UserCreateOpts{
    Name:   "deploy",
    Shell:  "/bin/bash",
    Groups: []string{"sudo", "docker"},
})

// Update a user (lock account)
lock := true
resp, err := c.User.Update(ctx, "web-01", "deploy", client.UserUpdateOpts{
    Lock: &lock,
})

// Change password
resp, err := c.User.ChangePassword(ctx, "web-01", "deploy", "newpass123")

// Delete a user
resp, err := c.User.Delete(ctx, "web-01", "deploy")

// List SSH keys for a user
keysResp, err := c.User.ListKeys(ctx, "web-01", "deploy")
for _, r := range keysResp.Data.Results {
    for _, k := range r.Keys {
        fmt.Printf("%s %s %s\n", k.Type, k.Fingerprint, k.Comment)
    }
}

// Add an SSH key
addResp, err := c.User.AddKey(ctx, "web-01", "deploy", client.SSHKeyAddOpts{
    Key: "ssh-ed25519 AAAA... user@laptop",
})

// Remove an SSH key by fingerprint
removeResp, err := c.User.RemoveKey(ctx, "web-01", "deploy", "SHA256:abc123...")
```

## Example

See
[`examples/sdk/client/user.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/user.go)
for a complete working example.

## SSH Key Types

### `SSHKeyInfoResult`

| Field      | Type           | Description             |
| ---------- | -------------- | ----------------------- |
| `Hostname` | `string`       | Target hostname         |
| `Status`   | `string`       | Operation status        |
| `Keys`     | `[]SSHKeyInfo` | List of authorized keys |
| `Error`    | `string`       | Error message (if any)  |

### `SSHKeyInfo`

| Field         | Type     | Description                    |
| ------------- | -------- | ------------------------------ |
| `Type`        | `string` | Key type (e.g., `ssh-ed25519`) |
| `Fingerprint` | `string` | SHA256 fingerprint             |
| `Comment`     | `string` | Key comment                    |

### `SSHKeyMutationResult`

| Field      | Type     | Description                         |
| ---------- | -------- | ----------------------------------- |
| `Hostname` | `string` | Target hostname                     |
| `Status`   | `string` | Operation status                    |
| `Changed`  | `bool`   | Whether the operation changed state |
| `Error`    | `string` | Error message (if any)              |

### `SSHKeyAddOpts`

| Field | Type     | Description                                            |
| ----- | -------- | ------------------------------------------------------ |
| `Key` | `string` | Full SSH public key line (e.g., `ssh-ed25519 AAAA...`) |

## Permissions

Requires `user:read` for List, Get, and ListKeys. Create, Update, Delete,
ChangePassword, AddKey, and RemoveKey require `user:write`.
