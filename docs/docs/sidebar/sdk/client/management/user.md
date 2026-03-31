---
sidebar_position: 5
---

# User

User account management on target hosts.

## Methods

| Method                                          | Description              |
| ----------------------------------------------- | ------------------------ |
| `List(ctx, hostname)`                           | List all user accounts   |
| `Get(ctx, hostname, name)`                      | Get a user by name       |
| `Create(ctx, hostname, opts)`                   | Create a user account    |
| `Update(ctx, hostname, name, opts)`             | Update a user account    |
| `Delete(ctx, hostname, name)`                   | Delete a user account    |
| `ChangePassword(ctx, hostname, name, password)` | Change a user's password |

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
```

## Example

See
[`examples/sdk/client/user.go`](https://github.com/retr0h/osapi/blob/main/examples/sdk/client/user.go)
for a complete working example.

## Permissions

Requires `user:read` for List and Get. Create, Update, Delete, and
ChangePassword require `user:write`.
