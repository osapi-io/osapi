---
sidebar_position: 6
---

# Group

Group management on target hosts.

## Methods

| Method                              | Description         |
| ----------------------------------- | ------------------- |
| `List(ctx, hostname)`               | List all groups     |
| `Get(ctx, hostname, name)`          | Get a group by name |
| `Create(ctx, hostname, opts)`       | Create a group      |
| `Update(ctx, hostname, name, opts)` | Update a group      |
| `Delete(ctx, hostname, name)`       | Delete a group      |

## Usage

```go
import "github.com/osapi-io/osapi/pkg/sdk/client"

c := client.New("http://localhost:8080", token)

// List all groups
resp, err := c.Group.List(ctx, "_any")
// Each result is one group on one host.
for _, r := range resp.Data.Results {
    fmt.Printf("%s: %s gid=%d\n", r.Hostname, r.Name, r.GID)
}

// Get a specific group
resp, err := c.Group.Get(ctx, "web-01", "docker")

// Create a group
resp, err := c.Group.Create(ctx, "web-01", client.GroupCreateOpts{
    Name: "deploy",
})

// Update group members
resp, err := c.Group.Update(ctx, "web-01", "deploy", client.GroupUpdateOpts{
    Members: []string{"alice", "bob"},
})

// Delete a group
resp, err := c.Group.Delete(ctx, "web-01", "deploy")
```

## Example

Refer to
[`examples/sdk/client/group.go`](https://github.com/osapi-io/osapi/blob/main/examples/sdk/client/group.go)
for a complete working example.

## Permissions

Requires `user:read` for List and Get. Create, Update, and Delete require
`user:write`.
